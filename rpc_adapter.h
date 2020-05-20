#ifndef A0_GO_RPC_ADAPTER_H
#define A0_GO_RPC_ADAPTER_H

#include <a0/rpc.h>
#include <a0/file_arena.h>

#include "common_adapter.h"
#include "packet_adapter.h"

extern void a0go_rpc_request_callback(void*, a0_rpc_request_t);

static inline errno_t a0go_rpc_server_init(a0_rpc_server_t* server,
                                           a0_buf_t arena,
                                           uintptr_t alloc_id,
                                           uintptr_t onrequest_id,
                                           uintptr_t oncancel_id) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  a0_rpc_request_callback_t onrequest = {
      .user_data = (void*)onrequest_id,
      .fn = a0go_rpc_request_callback,
  };
  a0_packet_id_callback_t oncancel = {
      .user_data = (void*)oncancel_id,
      .fn = a0go_packet_id_callback,
  };
  return a0_rpc_server_init(server, arena, alloc, onrequest, oncancel);
}

static inline errno_t a0go_rpc_server_async_close(a0_rpc_server_t* server, uintptr_t callback_id) {
  a0_callback_t callback = {
      .user_data = (void*)callback_id,
      .fn = a0go_callback,
  };
  return a0_rpc_server_async_close(server, callback);
}

static inline errno_t a0go_rpc_client_init(a0_rpc_client_t* client,
                                           a0_buf_t arena,
                                           uintptr_t alloc_id) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  return a0_rpc_client_init(client, arena, alloc);
}

static inline errno_t a0go_rpc_client_async_close(a0_rpc_client_t* client, uintptr_t callback_id) {
  a0_callback_t callback = {
      .user_data = (void*)callback_id,
      .fn = a0go_callback,
  };
  return a0_rpc_client_async_close(client, callback);
}

static inline errno_t a0go_rpc_send(a0_rpc_client_t* client,
                                    a0_packet_t pkt,
                                    uintptr_t packet_callback_id) {
  a0_packet_callback_t packet_callback = {
      .user_data = (void*)packet_callback_id,
      .fn = a0go_packet_callback,
  };
  return a0_rpc_send(client, pkt, packet_callback);
}

#endif  // A0_GO_RPC_ADAPTER_H
