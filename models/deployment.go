package models

type Blockchain struct {
	Name          string            `mapstructure:"name" binding:"required"`
	Ledger        string            `mapstructure:"ledger" binding:"required"`
	Version       string            `mapstructure:"version" binding:"required"`
	Consensus     string            `mapstructure:"consensus" binding:"required"`
	Nodes         int               `mapstructure:"Nodes" binding:"required"`
	Orchestration string            `mapstructure:"orchestration" binding:"required"`
	Processors    map[string]string `mapstructure:"processors" binding:"required"`
	Subscribers   map[string]string `mapstructure:"subscribers,omitempty"`
}
