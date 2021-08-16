// MIT License

// Copyright (c) 2021 eraft dev group

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

#ifndef ERAFT_KV_CALLBACK_H_
#define ERAFT_KV_CALLBACK_H_

#include <eraftio/raft_cmdpb.pb.h>

#include <atomic>

namespace kvserver {

struct Callback {
  Callback() {
    this->done_ = false;
    this->resp_ = nullptr;
  }

  void Done(raft_cmdpb::RaftCmdResponse* resp) {
    if (resp != nullptr) {
      this->resp_ = resp;
    }
    this->done_ = true;
  }

  raft_cmdpb::RaftCmdResponse* WaitResp() {
    if (this->done_) {
      return resp_;
    }
    return nullptr;
  }

  raft_cmdpb::RaftCmdResponse* WaitRespWithTimeout() {
    if (this->done_) {
    }
  }

  raft_cmdpb::RaftCmdResponse* resp_;

  std::atomic<bool> done_;
};

}  // namespace kvserver

#endif
