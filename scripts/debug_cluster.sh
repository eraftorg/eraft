docker run -it \
                --net eraft-network \
                --name eraft-1 \
                --hostname eraft-1 \
                --ip 172.19.0.11 \
                -v ${PWD}:/eraft \
                -w /eraft \
                -p 20160:20160 \
                eraft/eraft_pmem:v2
