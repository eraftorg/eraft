#ifndef ERAFT_KV_NODE_H_
#define ERAFT_KV_NODE_H_

#include <stdint.h>
#include <eraftio/metapb.pb.h>

#include <Kv/Config.h>
#include <Kv/RaftStore.h>
#include <Kv/Engines.h>
#include <Kv/Transport.h>

namespace kvserver
{

class Node
{

public:
    
    Node(/* args */);

    Node(RaftStore* system, Config* cfg);

    bool Start(Engines* engines, Transport trans);

    bool CheckStore(Engines& engs, uint64_t* storeId);

    bool BootstrapStore(Engines& engs, uint64_t* storeId);

    bool StartNode(Engines* engs, Transport trans);

    bool StopNode(uint64_t storeID);

    void Stop();

    uint64_t GetStoreID();

    ~Node();

private:
    
    uint64_t clusterID_;

    metapb::Store* store_;

    Config* cfg_;

    RaftStore* system_;
};

    
} // namespace kvserver


#endif