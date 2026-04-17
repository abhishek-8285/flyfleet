/**
 * High-performance RingBuffer using SharedArrayBuffer and Atomics.
 * Designed for zero-copy binary telemetry ingestion between workers.
 */
export class RingBuffer {
  private head: Int32Array;
  private tail: Int32Array;
  private buffer: Uint8Array;
  private size: number;
  private isShared: boolean;

  constructor(sharedBuffer: SharedArrayBuffer | ArrayBuffer, offset: number, size: number) {
    // Layout: [head (4 bytes), tail (4 bytes), data (size bytes)]
    this.isShared = typeof SharedArrayBuffer !== "undefined" && sharedBuffer instanceof SharedArrayBuffer;
    
    this.head = new Int32Array(sharedBuffer, offset, 1);
    this.tail = new Int32Array(sharedBuffer, offset + 4, 1);
    this.buffer = new Uint8Array(sharedBuffer, offset + 8, size);
    this.size = size;

    if (!this.isShared) {
      console.warn("[RingBuffer] Initialized in NON-SHARED mode. Performance will be degraded.");
    }
  }

  /**
   * Pushes binary data into the buffer.
   * Returns true if successful, false if buffer is full.
   */
  push(data: Uint8Array): boolean {
    const head = this.isShared ? Atomics.load(this.head, 0) : this.head[0];
    const tail = this.isShared ? Atomics.load(this.tail, 0) : this.tail[0];
    const available = (tail - head - 1 + this.size) % this.size;

    if (available < data.length) {
      return false; // Buffer overflow
    }

    for (let i = 0; i < data.length; i++) {
      this.buffer[(head + i) % this.size] = data[i];
    }

    if (this.isShared) {
      Atomics.store(this.head, 0, (head + data.length) % this.size);
    } else {
      this.head[0] = (head + data.length) % this.size;
    }
    return true;
  }

  /**
   * Pops data of specified length from the buffer.
   * Returns Uint8Array or null if insufficient data.
   */
  pop(length: number): Uint8Array | null {
    const head = this.isShared ? Atomics.load(this.head, 0) : this.head[0];
    const tail = this.isShared ? Atomics.load(this.tail, 0) : this.tail[0];
    const occupied = (head - tail + this.size) % this.size;

    if (occupied < length) {
      return null;
    }

    const data = new Uint8Array(length);
    for (let i = 0; i < length; i++) {
      data[i] = this.buffer[(tail + i) % this.size];
    }

    if (this.isShared) {
      Atomics.store(this.tail, 0, (tail + length) % this.size);
    } else {
      this.tail[0] = (tail + length) % this.size;
    }
    return data;
  }

  getOccupied(): number {
    const head = this.isShared ? Atomics.load(this.head, 0) : this.head[0];
    const tail = this.isShared ? Atomics.load(this.tail, 0) : this.tail[0];
    return (head - tail + this.size) % this.size;
  }
}
