import { CustomLayerInterface, Map } from "maplibre-gl";

/**
 * WebGL PointCloudLayer optimized for rendering 500k-1M entities as simple points.
 * Part of the LOD 1 Rendering Strategy (Mid-Zoom).
 */
export class PointCloudLayer implements CustomLayerInterface {
  id: string;
  type: "custom" = "custom";
  renderingMode: "2d" | "3d" = "2d";

  private program: WebGLProgram | null = null;
  private shards: Float32Array[];
  private vbos: WebGLBuffer[] = [];

  // [lng, lat, heading, speed, status, timestamp, ...reserved]
  private static STRIDE = 8;

  private focusedId: number = -1.0;

  constructor(id: string, shards: SharedArrayBuffer[]) {
    this.id = id;
    this.shards = shards.map((sab) => new Float32Array(sab));
  }

  setFocusedVehicle(pickingId: number | null) {
    this.focusedId = pickingId ?? -1.0;
  }

  onAdd(map: Map, gl: WebGLRenderingContext): void {
    const vertexSource = `
      uniform mat4 u_matrix;
      uniform float u_focused_id;
      attribute vec2 a_pos;
      attribute float a_time;
      attribute float a_picking_id;
      varying float v_focused;
      varying float v_active;
      
      void main() {
        v_active = a_time > 0.0 ? 1.0 : 0.0;
        v_focused = abs(a_picking_id - u_focused_id) < 1.0 ? 1.0 : 0.0;
        gl_Position = u_matrix * vec4(a_pos, 0.0, 1.0);
        gl_PointSize = v_focused > 0.5 ? 8.0 : 3.0; 
      }
    `;

    const fragmentSource = `
      precision mediump float;
      varying float v_active;
      varying float v_focused;
      void main() {
        if (v_active < 0.5) discard;
        if (v_focused > 0.5) {
            gl_FragColor = vec4(1.0, 0.4, 0.0, 1.0); // Orange
        } else {
            gl_FragColor = vec4(0.0, 0.8, 1.0, 0.8); // Blue
        }
      }
    `;

    this.program = this.createProgram(gl, vertexSource, fragmentSource);

    // Create VBOs for each shard
    this.vbos = [];
    this.shards.forEach(() => {
      this.vbos.push(gl.createBuffer()!);
    });
  }

  render(gl: WebGLRenderingContext, options: any): void {
    if (!this.program) return;

    gl.useProgram(this.program);
    gl.uniformMatrix4fv(
      gl.getUniformLocation(this.program, "u_matrix"),
      false,
      options.matrix,
    );

    const posLoc = gl.getAttribLocation(this.program, "a_pos");
    const timeLoc = gl.getAttribLocation(this.program, "a_time");
    const pickLoc = gl.getAttribLocation(this.program, "a_picking_id");
    const focusLoc = gl.getUniformLocation(this.program, "u_focused_id");

    gl.uniform1f(focusLoc, this.focusedId);

    this.vbos.forEach((vbo, i) => {
      gl.bindBuffer(gl.ARRAY_BUFFER, vbo);

      // Strategy: Only upload if dirty. In a real system, we'd use bitfields.
      // For now, we perform a partial sub-upload of the shard.
      gl.bufferData(gl.ARRAY_BUFFER, this.shards[i], gl.DYNAMIC_DRAW);

      gl.enableVertexAttribArray(posLoc);
      // We read from the same SAB as InstancedVehicleLayer
      // [prev_lng, prev_lat, prev_hdg, curr_lng, curr_lat, curr_hdg, start_time, picking_id]
      // PointCloud reads curr_lng (idx 3) and curr_lat (idx 4)
      gl.vertexAttribPointer(posLoc, 2, gl.FLOAT, false, 32, 12);

      gl.enableVertexAttribArray(timeLoc);
      gl.vertexAttribPointer(timeLoc, 1, gl.FLOAT, false, 32, 24);

      if (pickLoc !== -1) {
        gl.enableVertexAttribArray(pickLoc);
        gl.vertexAttribPointer(pickLoc, 1, gl.FLOAT, false, 32, 28);
      }

      gl.drawArrays(gl.POINTS, 0, this.shards[i].length / 8);
    });
  }

  private createProgram(
    gl: WebGLRenderingContext,
    vs: string,
    fs: string,
  ): WebGLProgram {
    const program = gl.createProgram()!;
    const vShader = gl.createShader(gl.VERTEX_SHADER)!;
    gl.shaderSource(vShader, vs);
    gl.compileShader(vShader);

    const fShader = gl.createShader(gl.FRAGMENT_SHADER)!;
    gl.shaderSource(fShader, fs);
    gl.compileShader(fShader);

    gl.attachShader(program, vShader);
    gl.attachShader(program, fShader);
    gl.linkProgram(program);
    return program;
  }
}
