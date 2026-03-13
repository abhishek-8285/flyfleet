/**
 * BitsetSlotManager: A memory-ultra-efficient manager for 1M+ entity slots.
 * Uses a Uint32Array as a bitmask to track availability.
 * Memory footprint for 1M slots: ~125 KB (vs 4MB for Int32Array free-list).
 */
export class SlotManager {
  private bits: Uint32Array;
  private lastSearchIndex: number = 0;
  public readonly maxSlots: number;

  constructor(maxSlots: number) {
    this.maxSlots = maxSlots;
    // 1 bit per slot, 32 bits per element
    this.bits = new Uint32Array(Math.ceil(maxSlots / 32));
    // Initialize all bits to 0 (all slots free)
  }

  /**
   * Helper to decompose a global index into shard and local index.
   */
  static decompose(
    index: number,
    shardSize: number,
  ): { shard: number; local: number } {
    return {
      shard: Math.floor(index / shardSize),
      local: index % shardSize,
    };
  }

  /**
   * Allocates a new slot index.
   * Returns -1 if no slots available.
   * Uses bit manipulation to find the first zero bit starting from lastSearchIndex.
   */
  allocate(): number {
    const totalElements = this.bits.length;
    let elementIndex = Math.floor(this.lastSearchIndex / 32);

    // Search for a word with at least one free slot (bit == 0)
    for (let i = 0; i < totalElements; i++) {
      const currentIdx = (elementIndex + i) % totalElements;
      const word = this.bits[currentIdx];

      if (word !== 0xffffffff) {
        // Word has at least one 0
        // Find the first 0 bit in the word
        for (let bit = 0; bit < 32; bit++) {
          if (!(word & (1 << bit))) {
            const finalIndex = currentIdx * 32 + bit;
            if (finalIndex < this.maxSlots) {
              this.bits[currentIdx] |= 1 << bit; // Set bit to 1 (occupied)
              this.lastSearchIndex = finalIndex;
              return finalIndex;
            }
          }
        }
      }
    }

    return -1;
  }

  /**
   * Returns a slot to the free pool.
   */
  release(index: number): void {
    if (index < 0 || index >= this.maxSlots) return;
    const elementIndex = Math.floor(index / 32);
    const bit = index % 32;
    this.bits[elementIndex] &= ~(1 << bit); // Clear bit to 0 (free)

    // Optimization: start next search here
    if (index < this.lastSearchIndex) {
      this.lastSearchIndex = index;
    }
  }

  getAvailableCount(): number {
    let occupied = 0;
    for (let i = 0; i < this.bits.length; i++) {
      let word = this.bits[i];
      // Brian Kernighan's bit-counting algorithm
      while (word) {
        word &= word - 1;
        occupied++;
      }
    }
    return Math.max(0, this.maxSlots - occupied);
  }
}
