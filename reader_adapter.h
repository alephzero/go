#ifndef A0_GO_READER_ADAPTER_H
#define A0_GO_READER_ADAPTER_H

#include <a0/err.h>
#include <a0/inline.h>
#include <a0/reader.h>

#include "alloc_adapter.h"
#include "packet_adapter.h"

A0_STATIC_INLINE
a0_err_t a0go_reader_sync_init(a0_reader_sync_t* reader_sync,
                               a0_arena_t arena,
                               uintptr_t alloc_id,
                               a0_reader_init_t init,
                               a0_reader_iter_t iter) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  return a0_reader_sync_init(reader_sync, arena, alloc, init, iter);
}

A0_STATIC_INLINE
a0_err_t a0go_reader_init(a0_reader_t* reader,
                          a0_arena_t arena,
                          uintptr_t alloc_id,
                          a0_reader_init_t init,
                          a0_reader_iter_t iter,
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
  return a0_reader_init(reader, arena, alloc, init, iter, packet_callback);
}

A0_STATIC_INLINE
a0_err_t a0go_reader_read_one(a0_arena_t arena,
                              uintptr_t alloc_id,
                              a0_reader_init_t init,
                              int flags,
                              a0_packet_t* out) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  return a0_reader_read_one(arena, alloc, init, flags, out);
}

#endif  // A0_GO_READER_ADAPTER_H
