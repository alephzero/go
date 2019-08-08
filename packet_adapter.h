#ifndef A0_GO_PACKET_ADAPTER_H
#define A0_GO_PACKET_ADAPTER_H

#include <a0/packet.h>
#include <stdlib.h>

#include "common_adapter.h"

static inline errno_t a0go_packet_build(size_t num_headers,
                                        a0_packet_header_t* headers,
                                        a0_buf_t payload,
                                        int alloc_id,
                                        a0_packet_t* out) {
  a0_alloc_t alloc = {
      .user_data = &alloc_id,
      .fn = a0go_alloc,
  };
  return a0_packet_build(num_headers, headers, payload, alloc, out);
}

#endif  // A0_GO_PACKET_ADAPTER_H
