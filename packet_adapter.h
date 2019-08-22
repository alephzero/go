#ifndef A0_GO_PACKET_ADAPTER_H
#define A0_GO_PACKET_ADAPTER_H

#include <a0/packet.h>

#include "common_adapter.h"

static inline errno_t a0go_packet_build(size_t num_headers,
                                        a0_packet_header_t* headers,
                                        a0_buf_t payload,
                                        uintptr_t alloc_id,
                                        a0_packet_t* out) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .fn = a0go_alloc,
  };
  return a0_packet_build(num_headers, headers, payload, alloc, out);
}

extern void a0go_packet_callback(void*, a0_packet_t);
extern void a0go_packet_id_callback(void*, char*);

#endif  // A0_GO_PACKET_ADAPTER_H
