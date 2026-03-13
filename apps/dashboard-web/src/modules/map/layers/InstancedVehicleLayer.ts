import { CustomLayerInterface, Map } from "maplibre-gl";

/**
 * Enterprise-grade WebGL Instanced Layer.
 * Supports:
 * 1. Dual-state temporal interpolation (60FPS smooth motion)
 * 2. GPU-based picking buffer interaction
 * 3. High-precision Mercator projection
 */
export class InstancedVehicleLayer implements CustomLayerInterface {
  id = "instanced-vehicles";
  type: "custom" = "custom";
  renderingMode: "2d" | "3d" = "2d";
  private matrixLoc: WebGLUniformLocation | null = null;
  private timeLoc: WebGLUniformLocation | null = null;
  private zoomLoc: WebGLUniformLocation | null = null;
  private pickingLoc: WebGLUniformLocation | null = null;

  private program: WebGLProgram | null = null;
  private vertexBuffer: WebGLBuffer | null = null;
  private instanceBuffers: (WebGLBuffer | null)[] = [];
  private shardCounts: number[] = [];
  private isInitialized: boolean[] = [];

  private focusIds: Set<number> = new Set();
  private focusUniformLoc: WebGLUniformLocation | null = null;

  constructor(private telemetryShards: Float32Array[]) {
    this.shardCounts = telemetryShards.map(() => 0);
    this.isInitialized = telemetryShards.map(() => false);
  }

  onAdd(map: Map, gl: WebGLRenderingContext) {
    const vertexSource = `
      uniform mat4 u_matrix;
      uniform float u_time;
      uniform float u_zoom;
      uniform bool u_picking;
      uniform float u_focused_id; // New uniform for focused ID

      attribute vec2 a_pos;
      // Instance attributes (Stride 8)
      attribute vec2 a_prev_pos;   // [0,1]
      attribute float a_prev_hdg;  // [2]
      attribute vec2 a_curr_pos;   // [3,4]
      attribute float a_curr_hdg;  // [5]
      attribute float a_start_time; // [6]
      attribute float a_picking_id; // [7]

      varying vec4 v_color;

      // High-precision Mercator projection constants
      const float PI = 3.141592653589793;
      
      vec2 project(vec2 lnglat) {
        float x = lnglat.x / 360.0 + 0.5;
        float sinLat = sin(lnglat.y * PI / 180.0);
        float y = 0.5 - 0.5 * log((1.0 + sinLat) / (1.0 - sinLat)) / (2.0 * PI);
        return vec2(x, y);
      }

      void main() {
        // 1. Calculate interpolation factor (0.0 to 1.0)
        // We assume a 1000ms transition period for telemetry batches
        float t = clamp((u_time - a_start_time) / 1000.0, 0.0, 1.0);
        
        // Easing (optional: easeInOutQuad)
        float easedT = t < 0.5 ? 2.0 * t * t : -1.0 + (4.0 - 2.0 * t) * t;

        vec2 pos = mix(a_prev_pos, a_curr_pos, easedT);
        float hdg = mix(a_prev_hdg, a_curr_hdg, easedT);

        // 2. Project to Mercator
        vec2 merchantPos = project(pos);
        
        // 3. Rotation logic
        float rad = -hdg * PI / 180.0;
        vec2 rotated = vec2(
          a_pos.x * cos(rad) - a_pos.y * sin(rad),
          a_pos.x * sin(rad) + a_pos.y * cos(rad)
        );

        // Scaling based on zoom to keep icons readable but precise
        float scale = 0.0001 * pow(2.0, 14.0 - u_zoom);
        
        // LOD Hardening: At global zoom levels, reduce icon size to avoid overlap noise
        if (u_zoom < 8.0) scale *= 0.5;
        if (u_zoom < 5.0) scale *= 0.2;
        
        gl_Position = u_matrix * vec4(merchantPos + rotated * scale, 0.0, 1.0);

        // Pass focus and speed to fragment shader
        v_focused = float(a_picking_id == u_focused_id);
        v_speed = a_speed;

        if (u_picking) {
          // Encode ID into RGBA
          float r = floor(a_picking_id / 65536.0) / 255.0;
          float g = floor(mod(a_picking_id, 65536.0) / 256.0) / 255.0;
          float b = mod(a_picking_id, 256.0) / 255.0;
          v_color = vec4(r, g, b, 1.0);
        } else {
          // This v_color is only used if u_picking is true.
          // The actual display color is determined by the fragment shader using v_focused and v_speed.
          v_color = vec4(0.14, 0.38, 0.92, 1.0); 
        }
      }
    `;

    const fragmentSource = `
      precision mediump float;
      varying vec4 v_color;
      void main() {
        gl_FragColor = v_color;
      }
    `;

    this.program = this.createProgram(gl, vertexSource, fragmentSource);
    this.matrixLoc = gl.getUniformLocation(this.program!, "u_matrix");
    this.timeLoc = gl.getUniformLocation(this.program!, "u_time");
    this.zoomLoc = gl.getUniformLocation(this.program!, "u_zoom");
    this.pickingLoc = gl.getUniformLocation(this.program!, "u_picking");

    this.vertexBuffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, this.vertexBuffer);
    const vertices = new Float32Array([0, 1, -0.6, -1, 0.6, -1]);
    gl.bufferData(gl.ARRAY_BUFFER, vertices, gl.STATIC_DRAW);

    this.instanceBuffers = this.telemetryShards.map(() => gl.createBuffer());
  }

  private createProgram(gl: WebGLRenderingContext, vs: string, fs: string) {
    const vShader = gl.createShader(gl.VERTEX_SHADER)!;
    gl.shaderSource(vShader, vs);
    gl.compileShader(vShader);
    if (!gl.getShaderParameter(vShader, gl.COMPILE_STATUS)) {
      console.error(gl.getShaderInfoLog(vShader));
    }

    const fShader = gl.createShader(gl.FRAGMENT_SHADER)!;
    gl.shaderSource(fShader, fs);
    gl.compileShader(fShader);

    const prog = gl.createProgram()!;
    gl.attachShader(prog, vShader);
    gl.attachShader(prog, fShader);
    gl.linkProgram(prog);
    return prog;
  }

  private dirtyIndicesByShard: Record<number, number[]> = {};

  updateData(shards: Float32Array[], globalDirtyIndices: number[]) {
    this.telemetryShards = shards;

    // Group dirty indices by shard for efficient bufferSubData calls
    this.dirtyIndicesByShard = {};
    const SHARD_SIZE = 100000;

    for (const index of globalDirtyIndices) {
      const shardIdx = Math.floor(index / SHARD_SIZE);
      const localIdx = index % SHARD_SIZE;

      if (!this.dirtyIndicesByShard[shardIdx]) {
        this.dirtyIndicesByShard[shardIdx] = [];
      }
      this.dirtyIndicesByShard[shardIdx].push(localIdx);
    }

    // Update shard counts
    this.shardCounts = shards.map((s) => s.length / 8);
  }

  setFocusedVehicle(pickingId: number | null) {
    this.focusedId = pickingId ?? -1.0;
  }

  private focusedId: number = -1.0;

  render(gl: WebGLRenderingContext, options: any) {
    if (!this.program || this.instanceBuffers.length === 0) return;

    const ext = gl.getExtension("ANGLE_instanced_arrays");
    if (!ext) return;

    gl.useProgram(this.program);

    // Set global uniforms once
    gl.uniformMatrix4fv(this.matrixLoc, false, options.matrix);
    gl.uniform1f(this.timeLoc, performance.now());
    gl.uniform1f(this.zoomLoc, options.zoom);
    gl.uniform1f(this.pickingLoc, options.isPicking ? 1.0 : 0.0); // Ensure picking uniform is set
    gl.uniform1f(this.focusUniformLoc, this.focusedId);

    // Loop through shards
    for (let i = 0; i < this.telemetryShards.length; i++) {
      const buffer = this.instanceBuffers[i];
      const shardData = this.telemetryShards[i];
      const shardCount = this.shardCounts[i];

      if (!buffer || shardCount === 0) continue;

      gl.bindBuffer(gl.ARRAY_BUFFER, buffer);

      // Upload dirty slots for this shard
      const localDirty = this.dirtyIndicesByShard[i];
      if (localDirty && localDirty.length > 0) {
        for (const localIdx of localDirty) {
          const offset = localIdx * 8 * 4;
          const data = shardData.subarray(localIdx * 8, (localIdx + 1) * 8);
          gl.bufferSubData(gl.ARRAY_BUFFER, offset, data);
        }
      } else if (!this.isInitialized[i]) {
        // Full initial upload if not initialized
        gl.bufferData(gl.ARRAY_BUFFER, shardData, gl.DYNAMIC_DRAW);
        this.isInitialized[i] = true;
      }

      // Setup attributes for this shard
      this.setupAttributes(gl, ext);

      // Draw this shard
      ext.drawArraysInstancedANGLE(gl.TRIANGLES, 0, 3, shardCount);
    }

    this.dirtyIndicesByShard = {};
  }

  private setupAttributes(gl: WebGLRenderingContext, ext: any) {
    const stride = 8 * 4;

    // Instanced Attributes
    this.setupAttribute(gl, ext, "a_prev_pos", 2, stride, 0);
    this.setupAttribute(gl, ext, "a_prev_hdg", 1, stride, 8);
    this.setupAttribute(gl, ext, "a_curr_pos", 2, stride, 12);
    this.setupAttribute(gl, ext, "a_curr_hdg", 1, stride, 20);
    this.setupAttribute(gl, ext, "a_start_time", 1, stride, 24);
    this.setupAttribute(gl, ext, "a_picking_id", 1, stride, 28);

    // Vertex Buffer Attributes (Triangle)
    gl.bindBuffer(gl.ARRAY_BUFFER, this.vertexBuffer);
    const posLoc = gl.getAttribLocation(this.program!, "a_pos");
    if (posLoc !== -1) {
      gl.enableVertexAttribArray(posLoc);
      gl.vertexAttribPointer(posLoc, 2, gl.FLOAT, false, 0, 0);
      ext.vertexAttribDivisorANGLE(posLoc, 0);
    }
  }

  private setupAttribute(
    gl: WebGLRenderingContext,
    ext: any,
    name: string,
    size: number,
    stride: number,
    offset: number,
  ) {
    const loc = gl.getAttribLocation(this.program!, name);
    if (loc !== -1) {
      gl.enableVertexAttribArray(loc);
      gl.vertexAttribPointer(loc, size, gl.FLOAT, false, stride, offset);
      ext.vertexAttribDivisorANGLE(loc, 1);
    }
  }
}
