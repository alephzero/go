#ifndef A0_GO_COMMON_ADAPTER_H
#define A0_GO_COMMON_ADAPTER_H

#include <a0/common.h>

extern void a0go_alloc(void*, size_t, a0_buf_t*);
extern void a0go_callback(void*);

// Utility to help copy Go pointers into C.
// TODO(lshamis): Maybe make dst a void**.
inline void a0go_copy_ptr(uint8_t** dst, uintptr_t src) {
  *dst = (uint8_t*)src;
}

#endif  // A0_GO_COMMON_ADAPTER_H
