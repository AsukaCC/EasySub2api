package service

// preferSchedulingAffinity applies affinity only after scheduling has established
// the candidates and their order. It replaces the first equivalent candidate;
// candidates in a better scheduling tier keep their positions. LRU/random order
// breaks ties only when no affinity is available in that tier.
func preferSchedulingAffinity[T any](ordered []T, stickyID string, accountOf func(T) *Account, sameTier func(T, T) bool) {
	if stickyID == "" {
		return
	}
	for i, candidate := range ordered {
		if accountOf(candidate).ID != stickyID {
			continue
		}
		for j := 0; j < i; j++ {
			if sameTier(ordered[j], candidate) {
				ordered[j], ordered[i] = ordered[i], ordered[j]
				return
			}
		}
		return
	}
}

func schedulingAccount(account *Account) *Account { return account }

func schedulingLoadAccount(item accountWithLoad) *Account { return item.account }

func sameAccountPriority(a, b *Account) bool { return a.Priority == b.Priority }

func sameAccountPriorityAndLoad(a, b accountWithLoad) bool {
	return a.account.Priority == b.account.Priority && a.loadInfo.LoadRate == b.loadInfo.LoadRate
}

func selectByAffinityOrLRU(candidates []accountWithLoad, stickyID string, preferOAuth bool) *accountWithLoad {
	for i := range candidates {
		if stickyID != "" && candidates[i].account.ID == stickyID {
			return &candidates[i]
		}
	}
	return selectByLRU(candidates, preferOAuth)
}
