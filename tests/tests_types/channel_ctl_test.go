package testtypes

import (
	"reflect"
	"testing"

	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	types "github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

func TestChannelCtl_DefaultChannels(t *testing.T) {
	ctl := types.NewChannelCtl[int]("test", nil)
	if ctl == nil {
		t.Fatalf("expected ctl instance")
	}

	chs := ctl.GetSubChannels()
	if len(chs) == 0 {
		t.Fatalf("expected default subchannels")
	}

	raw, typ, ok := ctl.GetSubChannelByName("ctl")
	if !ok || raw == nil || typ == nil {
		t.Fatalf("expected ctl subchannel and type")
	}
	if _, ok := raw.(ci.IChannelBase[any]); !ok {
		t.Fatalf("expected ctl to be an IChannelBase")
	}

	if _, ok := ctl.GetSubChannelTypeByName("done"); !ok {
		t.Fatalf("expected 'done' subchannel type present")
	}
	if buf, ok := ctl.GetSubChannelBuffersByName("condition"); !ok || buf <= 0 {
		t.Fatalf("expected condition buffers > 0")
	}
}

func TestChannelCtl_MainChannelAndClose(t *testing.T) {
	ctl := types.NewChannelCtl[string]("main", nil)
	if ctl == nil {
		t.Fatalf("expected ctl instance")
	}

	// Setar um main channel customizado
	ch := make(chan string, 3)

	mCh := ctl.SetMainChannel(ch)
	if mCh == nil {
		t.Fatalf("expected main channel set")
	}

	chM := ctl.GetMainChannel()
	if chM == nil {
		t.Fatalf("expected main channel get")
	}

	if typ := ctl.GetMainChannelType(); typ != reflect.TypeOf(ch) {
		t.Fatalf("unexpected main channel type: %v", typ)
	}

	if err := ctl.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}
}
