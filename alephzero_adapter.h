#ifndef A0_GO_ALEPHZERO_ADAPTER_H
#define A0_GO_ALEPHZERO_ADAPTER_H

#include <a0/alephzero.h>

#include "packet_adapter.h"

static inline errno_t a0go_config_reader_init(a0_subscriber_t* sub,
                                              a0_alephzero_t alephzero,
                                              uintptr_t packet_callback_id) {
  a0_packet_callback_t packet_callback = {
      .user_data = (void*)packet_callback_id,
      .fn = a0go_packet_callback,
  };
  return a0_config_reader_init(sub, alephzero, packet_callback);
}

static inline errno_t a0go_subscriber_init(a0_subscriber_t* sub,
                                           a0_alephzero_t alephzero,
                                           const char* name,
                                           a0_subscriber_read_start_t read_start,
                                           a0_subscriber_read_next_t read_next,
                                           uintptr_t packet_callback_id) {
  a0_packet_callback_t packet_callback = {
      .user_data = (void*)packet_callback_id,
      .fn = a0go_packet_callback,
  };
  return a0_subscriber_init(sub, alephzero, name, read_start, read_next, packet_callback);
}

static inline errno_t a0go_rpc_server_init(a0_rpc_server_t* server,
                                           a0_alephzero_t alephzero,
                                           const char* name,
                                           uintptr_t onrequest_id,
                                           uintptr_t oncancel_id) {
  a0_packet_callback_t onrequest = {
      .user_data = (void*)onrequest_id,
      .fn = a0go_packet_callback,
  };
  a0_packet_callback_t oncancel = {
      .user_data = (void*)oncancel_id,
      .fn = a0go_packet_callback,
  };
  return a0_rpc_server_init(server, alephzero, name, onrequest, oncancel);
}

#endif  // A0_GO_ALEPHZERO_ADAPTER_H
