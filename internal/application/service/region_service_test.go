package service

import "testing"

func TestRegionLabel(t *testing.T) {
	tests := []struct {
		region string
		want   string
	}{
		{region: "ap-east-2", want: "Asia Pacific (Taipei)"},
		{region: "ap-southeast-7", want: "Asia Pacific (Thailand)"},
		{region: "eusc-de-east-1", want: "European Sovereign Cloud (Brandenburg)"},
		{region: "future-region-1", want: "future-region-1"},
	}

	for _, test := range tests {
		t.Run(test.region, func(t *testing.T) {
			if got := RegionLabel(test.region); got != test.want {
				t.Fatalf("RegionLabel(%q) = %q, want %q", test.region, got, test.want)
			}
		})
	}
}

func TestRegionService_GetRegionCatalogReturnsCopy(t *testing.T) {
	service := NewRegionService(nil)
	first := service.GetRegionCatalog()
	if len(first) < 39 {
		t.Fatalf("GetRegionCatalog() returned %d regions, want at least 39", len(first))
	}

	first[0].Label = "changed"
	second := service.GetRegionCatalog()
	if second[0].Label == "changed" {
		t.Fatal("GetRegionCatalog() returned mutable catalog storage")
	}
}
