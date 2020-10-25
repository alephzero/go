#ifndef A0_GO_PRPC_ADAPTER_H
#define A0_GO_PRPC_ADAPTER_H

#include <a0/prpc.h>

#include <stdbool.h>

#include "alloc_adapter.h"
#include "common_adapter.h"
#include "packet_adapter.h"

extern void a0go_prpc_connection_callback(void*, a0_prpc_connection_t);
extern void a0go_prpc_callback(void*, a0_packet_t, bool);

static inline errno_t a0go_prpc_server_init(a0_prpc_server_t* server,
                                            a0_buf_t arena,
                                            uintptr_t alloc_id,
                                            uintptr_t onconnectId,
                                            uintptr_t oncancel_id) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  a0_prpc_connection_callback_t onconnect = {
      .user_data = (void*)onconnectId,
      .fn = a0go_prpc_connection_callback,
  };
  a0_packet_id_callback_t oncancel = {
      .user_data = (void*)oncancel_id,
      .fn = a0go_packet_id_callback,
  };
  return a0_prpc_server_init(server, arena, alloc, onconnect, oncancel);
}

static inline errno_t a0go_prpc_server_async_close(a0_prpc_server_t* server,
                                                   uintptr_t callback_id) {
  a0_callback_t callback = {
      .user_data = (void*)callback_id,
      .fn = a0go_callback,
  };
  return a0_prpc_server_async_close(server, callback);
}

static inline errno_t a0go_prpc_client_init(a0_prpc_client_t* client,
                                            a0_buf_t arena,
                                            uintptr_t alloc_id) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  return a0_prpc_client_init(client, arena, alloc);
}

static inline errno_t a0go_prpc_client_async_close(a0_prpc_client_t* client,
                                                   uintptr_t callback_id) {
  a0_callback_t callback = {
      .user_data = (void*)callback_id,
      .fn = a0go_callback,
  };
  return a0_prpc_client_async_close(client, callback);
}

static inline errno_t a0go_prpc_connect(a0_prpc_client_t* client,
                                        a0_packet_t pkt,
                                        uintptr_t prpc_callback_id) {
  a0_prpc_callback_t prpc_callback = {
      .user_data = (void*)prpc_callback_id,
      .fn = a0go_prpc_callback,
  };
  return a0_prpc_connect(client, pkt, prpc_callback);
}

#endif  // A0_GO_PRPC_ADAPTER_H
