iptables -t mangle -N GO_PROXY
iptables -t mangle -A GO_PROXY -p tcp -j TPROXY --on-port xxxx --tproxy-mark 0x1/0x1 
iptables -t mangle -A GO_PROXY -p udp -j TPROXY --on-port xxxx --tproxy-mark 0x1/0x1

iptables -t mangle -A PREROUTING -p tcp -i ${input interface} -j GO_PROXY 
iptables -t mangle -A PREROUTING -p udp -i ${input interface} -j GO_PROXY

ip rule add fwmark 0x1/0x1 lookup 100 ip route add local default dev lo table 100
