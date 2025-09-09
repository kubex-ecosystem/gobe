package gdbasez

import (
	tnl "github.com/kubex-ecosystem/gdbase/factory"
)

type TunnelMode = tnl.TunnelMode

const (
	TunnelQuick TunnelMode = "quick" // HTTP efêmero (URL dinâmica)
	TunnelNamed TunnelMode = "named" // HTTP+TCP fixo (Access)
)

type CloudflaredOpts = tnl.CloudflaredOpts
type TunnelHandle = tnl.TunnelHandle

func NewCloudflaredOpts(mode TunnelMode, networkName, targetDNS string, targetPort int, token string) CloudflaredOpts {
	return tnl.NewCloudflaredOpts(mode, networkName, targetDNS, targetPort, token)
}
