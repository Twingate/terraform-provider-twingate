package model

type SSHDownstream struct {
	Port int64
}

type SSHUpstream struct {
	Port int64
}

type SSHResource struct {
	PAMResource

	Downstream *SSHDownstream
	Upstream   *SSHUpstream
}
