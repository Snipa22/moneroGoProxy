package support

import "testing"

func TestMinerGetTarget(t *testing.T) {
	ut, st := uint32(4294967295), "ffffffff"
	uv, sv := getTarget(1)
	if uv != ut {
		t.Fatalf("Failed to get target uint correctly, expected %v got %v", ut, uv)
	}
	if sv != st {
		t.Fatalf("Failed to get target uint correctly, expected %v got %v", st, sv)
	}
	ut, st = uint32(4294967167), "ffffff7f"
	uv, sv = getTarget(2)
	if uv != ut {
		t.Fatalf("Failed to get target uint correctly, expected %v got %v", ut, uv)
	}
	if sv != st {
		t.Fatalf("Failed to get target uint correctly, expected %v got %v", st, sv)
	}
	ut, st = uint32(4294967040), "ffffff00"
	uv, sv = getTarget(256)
	if uv != ut {
		t.Fatalf("Failed to get target uint correctly, expected %v got %v", ut, uv)
	}
	if sv != st {
		t.Fatalf("Failed to get target uint correctly, expected %v got %v", st, sv)
	}
	ut, st = uint32(893911040), "35480000"
	uv, sv = getTarget(232342)
	if uv != ut {
		t.Fatalf("Failed to get target uint correctly, expected %v got %v", ut, uv)
	}
	if sv != st {
		t.Fatalf("Failed to get target uint correctly, expected %v got %v", st, sv)
	}
}
