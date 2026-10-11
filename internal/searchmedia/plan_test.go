package searchmedia

import "testing"

func TestPlanForRequest(t *testing.T) {
	cases := []struct {
		name      string
		requested []string
		limit     int
		want      map[string]int
	}{
		{"splits evenly", nil, 9, map[string]int{"klipy": 3, "google": 3, "giphy": 3}},
		{"remainder goes to the first available", nil, 10, map[string]int{"klipy": 4, "google": 3, "giphy": 3}},
		{"only requested services get a share", []string{"giphy", "klipy"}, 10, map[string]int{"klipy": 5, "google": 0, "giphy": 5}},
		{"requested order does not pick the remainder", []string{"giphy", "klipy"}, 5, map[string]int{"klipy": 3, "giphy": 2}},
		{"unavailable requested service is ignored", []string{"klipy", "imgur"}, 4, map[string]int{"klipy": 4}},
		{"no overlap gives every service zero", []string{"imgur"}, 10, map[string]int{"klipy": 0, "google": 0, "giphy": 0}},
		{"unknown service name gets zero", nil, 9, map[string]int{"imgur": 0}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plan := PlanForRequest(AvailableServiceNames, c.requested, c.limit)

			for service, want := range c.want {
				if got := plan.LimitFor(service); got != want {
					t.Errorf("%s: got %d, want %d", service, got, want)
				}
			}
		})
	}
}
