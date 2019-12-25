package types

// ResponseType acts as the JSON template for API response (successful calls)
type ResponseType struct {
	ValueType string      `json:"type"`
	Value     interface{} `json:"value"`
}

// ErrorType acts as the JSON template for API response (failed calls)
type ErrorType struct {
	Error string `json:"error"`
}

// KeyInfoType acts as the JSON template for element in KeyListType
type KeyInfoType struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

// KeyListType acts as the JSON template for API response (successful calls)
type KeyListType struct {
	Count int           `json:"count"`
	Total int           `json:"total"`
	Keys  []KeyInfoType `json:"keys"`
}

type InfoServer struct {
	RedisVersion    string `json:"redis_version"`
	RedisBuildId    string `json:"redis_build_id"`
	RedisMode       string `json:"redis_mode"`
	Os              string `json:"os"`
	ArchBits        string `json:"arch_bits"`
	GccVersion      string `json:"gcc_version"`
	ProcessId       string `json:"process_id"`
	RunId           string `json:"run_id"`
	TcpPort         string `json:"tcp_port"`
	UptimeInSeconds string `json:"uptime_in_seconds"`
	UptimeInDays    string `json:"uptime_in_days"`
	Hz              string `json:"hz"`
	ConfiguredHz    string `json:"configured_hz"`
	LruClock        string `json:"lru_clock"`
	Executable      string `json:"executable"`
	ConfigFile      string `json:"config_file"`
}

type InfoClients struct {
	ConnectedClients string `json:"connected_clients"`
	BlockedClients   string `json:"blocked_clients"`
}

type InfoReplication struct {
	Role            string `json:"role"`
	ConnectedSlaves string `json:"connected_slaves"`
	MasterReplId    string `json:"master_replid"`
	MasterReplId2   string `json:"master_replid2"`
}

type InfoCpu struct {
	UsedCpuSys          string `json:"used_cpu_sys"`
	UsedCpuUser         string `json:"used_cpu_user"`
	UsedCpuSysChildren  string `json:"used_cpu_sys_children"`
	UsedCpuUserChildren string `json:"used_cpu_user_children"`
}

type InfoCluster struct {
	ClusterEnabled string `json:"cluster_enabled"`
}

type Info struct {
	Server      InfoServer
	Clients     InfoClients
	Replication InfoReplication
	Cpu         InfoCpu
	Cluster     InfoCluster
}
