package searchmedia

import "slices"

// ServiceLimitPlan says how the requested limit is split between services.
type ServiceLimitPlan struct {
	limits map[string]int
}

// A nil requestedServiceNames means every available service. The remainder goes to the first services in available order, not in requested order.
func PlanForRequest(availableServiceNames, requestedServiceNames []string, limit int) ServiceLimitPlan {
	var names []string

	for _, name := range availableServiceNames {
		if requestedServiceNames == nil || slices.Contains(requestedServiceNames, name) {
			names = append(names, name)
		}
	}

	limits := make(map[string]int, len(names))

	if len(names) == 0 {
		return ServiceLimitPlan{limits: limits}
	}

	perService, remainder := limit/len(names), limit%len(names)

	for index, name := range names {
		limits[name] = perService
		// Add one extra item to the first services to distribute the remainder.
		if index < remainder {
			limits[name]++
		}
	}

	return ServiceLimitPlan{limits: limits}
}

func (p ServiceLimitPlan) LimitFor(serviceName string) int {
	return p.limits[serviceName]
}
