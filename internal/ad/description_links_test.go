package ad

import "testing"

func TestSplitDescriptionLinks(t *testing.T) {
	tests := []struct {
		in   string
		want []DescriptionSegment
	}{
		{
			in:   "plain text",
			want: []DescriptionSegment{{Text: "plain text"}},
		},
		{
			in: "See https://example.com for details.",
			want: []DescriptionSegment{
				{Text: "See "},
				{Link: &DescriptionLink{
					Text: "https://example.com",
					URL:  "https://example.com",
				}},
				{Text: " for details."},
			},
		},
		{
			in: "Visit https://a.test and https://b.test/path?q=1",
			want: []DescriptionSegment{
				{Text: "Visit "},
				{Link: &DescriptionLink{
					Text: "https://a.test",
					URL:  "https://a.test",
				}},
				{Text: " and "},
				{Link: &DescriptionLink{
					Text: "https://b.test/path?q=1",
					URL:  "https://b.test/path?q=1",
				}},
			},
		},
		{
			in:   "http://example.com stays plain",
			want: []DescriptionSegment{{Text: "http://example.com stays plain"}},
		},
		{
			in: "Details at (https://example.com).",
			want: []DescriptionSegment{
				{Text: "Details at ("},
				{Link: &DescriptionLink{
					Text: "https://example.com",
					URL:  "https://example.com",
				}},
				{Text: ")."},
			},
		},
		{
			in: "Specs at https://www.trek.com/us/en_US/bikes/road-bikes/endurance-road/domane/. Perfect",
			want: []DescriptionSegment{
				{Text: "Specs at "},
				{Link: &DescriptionLink{
					Text: "https://www.trek.com/us/en_US/bikes/road-bikes/endurance-road/domane/",
					URL:  "https://www.trek.com/us/en_US/bikes/road-bikes/endurance-road/domane/",
				}},
				{Text: "."},
				{Text: " Perfect"},
			},
		},
	}
	for _, tt := range tests {
		got := SplitDescriptionLinks(tt.in)
		if len(got) != len(tt.want) {
			t.Fatalf("SplitDescriptionLinks(%q) len = %d, want %d",
				tt.in, len(got), len(tt.want))
		}
		for i := range tt.want {
			if !descriptionSegmentEqual(got[i], tt.want[i]) {
				t.Errorf("SplitDescriptionLinks(%q)[%d] = %+v, want %+v",
					tt.in, i, got[i], tt.want[i])
			}
		}
	}
}

func descriptionSegmentEqual(a, b DescriptionSegment) bool {
	if a.Text != b.Text {
		return false
	}
	if (a.Link == nil) != (b.Link == nil) {
		return false
	}
	if a.Link == nil {
		return true
	}
	return a.Link.Text == b.Link.Text && a.Link.URL == b.Link.URL
}
