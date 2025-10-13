package config

// SovereignConfig holds sovereign config
type SovereignConfig struct {
	ExtendedShardHdrNonceHashStorage StorageConfig
	ExtendedShardHeaderStorage       StorageConfig
	MainChainNotarization            map[string]MainChainNotarization `toml:"MainChainNotarization"`
	OutgoingSubscribedEvents         OutgoingSubscribedEvents         `toml:"OutgoingSubscribedEvents"`
	OutGoingBridge                   OutGoingBridge                   `toml:"OutGoingBridge"`
	NotifierConfig                   NotifierConfig                   `toml:"NotifierConfig"`
	ETHNotifierConfig                ETHNotifierConfig                `toml:"ETHNotifierConfig"`
	SUINotifierConfig                SUINotifierConfig                `toml:"SUINotifierConfig"`
	GenesisConfig                    GenesisConfig                    `toml:"GenesisConfig"`
	OutGoingBridgeCertificate        OutGoingBridgeCertificate
}

// OutgoingSubscribedEvents holds config for outgoing subscribed events
type OutgoingSubscribedEvents struct {
	TimeToWaitForUnconfirmedOutGoingOperationInSeconds uint32            `toml:"TimeToWaitForUnconfirmedOutGoingOperationInSeconds"`
	SubscribedEvents                                   []SubscribedEvent `toml:"SubscribedEvents"`
}

// MainChainNotarization defines necessary data to start main chain notarization on a sovereign shard
type MainChainNotarization struct {
	StartRound uint64 `toml:"StartRound"`
}

// OutGoingBridge holds config for grpc client to send outgoing bridge txs
type OutGoingBridge struct {
	Enabled  bool   `toml:"Enabled"`
	GRPCHost string `toml:"GRPCHost"`
	GRPCPort string `toml:"GRPCPort"`
	Hasher   string `toml:"Hasher"`
}

// OutGoingBridgeCertificate holds config for outgoing bridge certificate paths
type OutGoingBridgeCertificate struct {
	CertificatePath   string
	CertificatePkPath string
}

// NotifierConfig holds sovereign notifier configuration
type NotifierConfig struct {
	Enabled                bool              `toml:"Enabled"`
	SubscribedEvents       []SubscribedEvent `toml:"SubscribedEvents"`
	WebSocketConfig        WebSocketConfig   `toml:"WebSocket"`
	AddressPubKeyConverter PubkeyConfig      `toml:"AddressPubKeyConverter"`
}

// ETHNotifierConfig holds eth notifier config
type ETHNotifierConfig struct {
	Enabled               bool                 `toml:"Enabled"`
	SubscribedEvents      []ETHSubscribedEvent `toml:"SubscribedEvents"`
	HasherType            string               `toml:"HasherType"`
	MarshallerType        string               `toml:"MarshallerType"`
	MinBlocksConfirmation uint8                `toml:"MinBlocksConfirmation"`
	BlockCacheSize        uint64               `toml:"BlockCacheSize"`
	URL                   string               `toml:"URL"`
}

// SubscribedEvent holds subscribed events config
type SubscribedEvent struct {
	Identifier string   `toml:"Identifier"`
	Addresses  []string `toml:"Addresses"`
}

// ETHSubscribedEvent holds eth subsribed events config
type ETHSubscribedEvent struct {
	Identifier string `toml:"Identifier"`
	Address    string `toml:"Address"`
}

// WebSocketConfig holds web socket config
type WebSocketConfig struct {
	Url                string `toml:"Url"`
	MarshallerType     string `toml:"MarshallerType"`
	RetryDuration      uint32 `toml:"RetryDuration"`
	BlockingAckOnError bool   `toml:"BlockingAckOnError"`
	HasherType         string `toml:"HasherType"`
	Mode               string `toml:"Mode"`
	WithAcknowledge    bool   `toml:"WithAcknowledge"`
	AcknowledgeTimeout int    `toml:"AcknowledgeTimeout"`
	Version            uint32 `toml:"Version"`
}

// SUINotifierConfig holds SUI notifier general config
type SUINotifierConfig struct {
	Enabled bool `toml:"Enabled"`

	MarshallerType string `toml:"MarshallerType"`
	HasherType     string `toml:"HasherType"`

	PoolingTime        uint8  `toml:"PoolingTime"`
	BatchSize          uint64 `toml:"BatchSize"`
	StartingCheckpoint uint64 `toml:"StartingCheckpoint"`

	SubscribedEvents []SUISubscribedEvent `toml:"SubscribedEvents"`
	ClientConfig     SUIClientConfig      `toml:"ClientConfig"`
}

// SUISubscribedEvent holds subscribed SUI events to be received via ws
type SUISubscribedEvent struct {
	EventType string `toml:"EventType"`
	Value     string `toml:"Value"`
}

// SUIClientConfig holds SUI client connection urls
type SUIClientConfig struct {
	RPCUrl string `toml:"RPCUrl"`
	WSUrl  string `toml:"WSUrl"`
}

// GenesisConfig should hold all sovereign genesis related configs
type GenesisConfig struct {
	NativeESDT string `toml:"NativeESDT"`
}
