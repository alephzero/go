#ifndef A0_GO_PUBSUB_ADAPTER_H
#define A0_GO_PUBSUB_ADAPTER_H

#include <a0/pubsub.h>

#include "alloc_adapter.h"
#include "common_adapter.h"
#include "packet_adapter.h"

static inline errno_t a0go_subscriber_sync_init(a0_subscriber_sync_t* sub_sync,
                                                a0_buf_t arena,
                                                uintptr_t alloc_id,
                                                a0_subscriber_init_t sub_init,
                                                a0_subscriber_iter_t sub_iter) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  return a0_subscriber_sync_init(sub_sync, arena, alloc, sub_init, sub_iter);
}

static inline errno_t a0go_subscriber_init(a0_subscriber_t* sub,
                                           a0_buf_t arena,
                                           uintptr_t alloc_id,
                                           a0_subscriber_init_t sub_init,
                                           a0_subscriber_iter_t sub_iter,
                                           uintptr_t packet_callback_id) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  a0_packet_callback_t packet_callback = {
      .user_data = (void*)packet_callback_id,
      .fn = a0go_packet_callback,
  };
  return a0_subscriber_init(sub, arena, alloc, sub_init, sub_iter, packet_callback);
}

static inline errno_t a0go_subscriber_async_close(a0_subscriber_t* sub, uintptr_t callback_id) {
  a0_callback_t callback = {
      .user_data = (void*)callback_id,
      .fn = a0go_callback,
  };
  return a0_subscriber_async_close(sub, callback);
}

static inline errno_t a0go_subscriber_read_one(a0_buf_t arena,
                                               uintptr_t alloc_id,
                                               a0_subscriber_init_t sub_init,
                                               int flags,
                                               a0_packet_t* out) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  return a0_subscriber_read_one(arena, alloc, sub_init, flags, out);
}

#endif  // A0_GO_PUBSUB_ADAPTER_H
