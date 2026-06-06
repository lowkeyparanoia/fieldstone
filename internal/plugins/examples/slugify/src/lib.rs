//! Fieldstone WASM example plugin: `slugify`.
//!
//! A tiny, dependency-free plugin compiled to `wasm32-unknown-unknown` that turns
//! an arbitrary UTF-8 string (e.g. a record's `title`) into a URL-safe slug:
//!
//!     "Shipping Realtime to Production!"  ->  "shipping-realtime-to-production"
//!
//! It matches the host ABI in `internal/plugins/wasm.go` exactly:
//!   exports `memory`, `malloc(size i32) -> i32`, and the plugin function
//!   `slugify(in_ptr, in_len, out_ptr_ptr, out_len_ptr) -> i32` which:
//!     1. reads `in_len` bytes at `in_ptr`,
//!     2. allocates an output buffer via `malloc`,
//!     3. writes the slug there,
//!     4. stores the output pointer + length (little-endian u32) into the two
//!        4-byte slots the host passed (`out_ptr_ptr`, `out_len_ptr`),
//!     5. returns 0 on success (non-zero = error).
//!
//! Build:  see build.sh  (rustc --target wasm32-unknown-unknown --crate-type cdylib …)

#![no_std]

use core::panic::PanicInfo;
use core::ptr;

#[panic_handler]
fn panic(_: &PanicInfo) -> ! {
    // Plugins must never unwind into the host; abort the loop instead.
    loop {}
}

// ---- Simple bump allocator over a fixed arena in linear memory ----
// The host calls `malloc` a few small times per invocation (input buffer, two
// 4-byte pointer slots, output buffer). An 8 MiB arena comfortably covers many
// invocations of this short-lived, single-shot plugin.
const ARENA_SIZE: usize = 8 * 1024 * 1024;
static mut ARENA: [u8; ARENA_SIZE] = [0u8; ARENA_SIZE];
static mut BUMP: usize = 0;

/// Allocate `size` bytes from the arena, 8-byte aligned. Returns a linear-memory
/// address (offset), or 0 if the arena is exhausted.
#[no_mangle]
pub extern "C" fn malloc(size: i32) -> i32 {
    let s = if size < 0 { 0 } else { size as usize };
    unsafe {
        let mut off = ptr::read(ptr::addr_of!(BUMP));
        off = (off + 7) & !7usize; // align to 8
        if off + s > ARENA_SIZE {
            return 0;
        }
        ptr::write(ptr::addr_of_mut!(BUMP), off + s);
        let base = ptr::addr_of_mut!(ARENA) as *mut u8;
        (base as usize + off) as i32
    }
}

/// Map a byte to a lowercase ASCII alphanumeric, or 0 if it is a separator.
fn lower_alnum(b: u8) -> u8 {
    match b {
        b'A'..=b'Z' => b + 32,
        b'a'..=b'z' => b,
        b'0'..=b'9' => b,
        _ => 0,
    }
}

unsafe fn write_u32_le(addr: usize, v: u32) {
    let p = addr as *mut u8;
    *p.add(0) = (v & 0xff) as u8;
    *p.add(1) = ((v >> 8) & 0xff) as u8;
    *p.add(2) = ((v >> 16) & 0xff) as u8;
    *p.add(3) = ((v >> 24) & 0xff) as u8;
}

/// Slugify the input string. See module docs for the ABI contract.
#[no_mangle]
pub extern "C" fn slugify(in_ptr: i32, in_len: i32, out_ptr_ptr: i32, out_len_ptr: i32) -> i32 {
    let n = if in_len < 0 { 0 } else { in_len as usize };
    let cap = if n == 0 { 1 } else { n }; // slug is never longer than the input
    let out_addr = malloc(cap as i32);
    if out_addr == 0 {
        return 1; // out of memory
    }

    unsafe {
        let inp = (in_ptr as u32 as usize) as *const u8;
        let out = (out_addr as u32 as usize) as *mut u8;

        let mut j: usize = 0;
        let mut prev_dash = false;
        let mut i: usize = 0;
        while i < n {
            let b = *inp.add(i);
            i += 1;
            let lc = lower_alnum(b);
            if lc != 0 {
                *out.add(j) = lc;
                j += 1;
                prev_dash = false;
            } else if j > 0 && !prev_dash {
                *out.add(j) = b'-';
                j += 1;
                prev_dash = true;
            }
        }
        // Trim a trailing separator.
        while j > 0 && *out.add(j - 1) == b'-' {
            j -= 1;
        }

        write_u32_le(out_ptr_ptr as u32 as usize, out_addr as u32);
        write_u32_le(out_len_ptr as u32 as usize, j as u32);
    }
    0
}
