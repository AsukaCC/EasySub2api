package domain

import (
	"strings"
	"time"

	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

const (
	AnnouncementStatusDraft    = "draft"
	AnnouncementStatusActive   = "active"
	AnnouncementStatusArchived = "archived"
)

const (
	AnnouncementNotifyModeSilent = "silent"
	AnnouncementNotifyModePopup  = "popup"
	AnnouncementNotifyModeBanner = "banner"
)

const (
	AnnouncementConditionTypeSubscription = "subscription"
	AnnouncementConditionTypeBalance      = "balance"
	AnnouncementConditionTypeUser         = "user"
	AnnouncementConditionTypeLevel        = "level"
)

const (
	AnnouncementOperatorIn  = "in"
	AnnouncementOperatorGT  = "gt"
	AnnouncementOperatorGTE = "gte"
	AnnouncementOperatorLT  = "lt"
	AnnouncementOperatorLTE = "lte"
	AnnouncementOperatorEQ  = "eq"
)

var (
	ErrAnnouncementNotFound      = infraerrors.NotFound("ANNOUNCEMENT_NOT_FOUND", "announcement not found")
	ErrAnnouncementInvalidTarget = infraerrors.BadRequest("ANNOUNCEMENT_INVALID_TARGET", "invalid announcement targeting rules")
)

type AnnouncementTargeting struct {
	// AnyOf 表示 OR：任意一个条件组满足即可展示。
	AnyOf []AnnouncementConditionGroup `json:"any_of,omitempty"`
}

type AnnouncementConditionGroup struct {
	// AllOf 表示 AND：组内所有条件都满足才算命中该组。
	AllOf []AnnouncementCondition `json:"all_of,omitempty"`
}

type AnnouncementCondition struct {
	// Type: subscription | balance | user | level
	Type string `json:"type"`

	// Operator:
	// - subscription/user/level: in
	// - balance: gt/gte/lt/lte/eq
	Operator string `json:"operator"`

	// subscription 条件：匹配的订阅套餐（group_id）
	GroupIDs []string `json:"group_ids,omitempty"`

	// balance 条件：比较阈值
	Value float64 `json:"value,omitempty"`

	// user 条件：匹配的用户 ID
	UserIDs []string `json:"user_ids,omitempty"`

	// level 条件：匹配的消费等级（1-3）
	Levels []int `json:"levels,omitempty"`
}

func (t AnnouncementTargeting) Matches(balance float64, activeSubscriptionGroupIDs map[string]struct{}) bool {
	return t.MatchesForUser(balance, activeSubscriptionGroupIDs, "", 0)
}

func (t AnnouncementTargeting) MatchesForUser(
	balance float64,
	activeSubscriptionGroupIDs map[string]struct{},
	userID string,
	userLevel int,
) bool {
	// 空规则：展示给所有用户
	if len(t.AnyOf) == 0 {
		return true
	}

	for _, group := range t.AnyOf {
		if len(group.AllOf) == 0 {
			// 空条件组不命中（避免 OR 中出现无条件 “全命中”）
			continue
		}
		allMatched := true
		for _, cond := range group.AllOf {
			if !cond.MatchesForUser(balance, activeSubscriptionGroupIDs, userID, userLevel) {
				allMatched = false
				break
			}
		}
		if allMatched {
			return true
		}
	}

	return false
}

func (c AnnouncementCondition) Matches(balance float64, activeSubscriptionGroupIDs map[string]struct{}) bool {
	return c.MatchesForUser(balance, activeSubscriptionGroupIDs, "", 0)
}

func (c AnnouncementCondition) MatchesForUser(
	balance float64,
	activeSubscriptionGroupIDs map[string]struct{},
	userID string,
	userLevel int,
) bool {
	switch c.Type {
	case AnnouncementConditionTypeSubscription:
		if c.Operator != AnnouncementOperatorIn {
			return false
		}
		if len(c.GroupIDs) == 0 {
			return false
		}
		if len(activeSubscriptionGroupIDs) == 0 {
			return false
		}
		for _, gid := range c.GroupIDs {
			if _, ok := activeSubscriptionGroupIDs[gid]; ok {
				return true
			}
		}
		return false

	case AnnouncementConditionTypeBalance:
		switch c.Operator {
		case AnnouncementOperatorGT:
			return balance > c.Value
		case AnnouncementOperatorGTE:
			return balance >= c.Value
		case AnnouncementOperatorLT:
			return balance < c.Value
		case AnnouncementOperatorLTE:
			return balance <= c.Value
		case AnnouncementOperatorEQ:
			return balance == c.Value
		default:
			return false
		}

	case AnnouncementConditionTypeUser:
		if c.Operator != AnnouncementOperatorIn || userID == "" {
			return false
		}
		for _, id := range c.UserIDs {
			if id == userID {
				return true
			}
		}
		return false

	case AnnouncementConditionTypeLevel:
		if c.Operator != AnnouncementOperatorIn || userLevel <= 0 {
			return false
		}
		for _, level := range c.Levels {
			if level == userLevel {
				return true
			}
		}
		return false

	default:
		return false
	}
}

func (t AnnouncementTargeting) NormalizeAndValidate() (AnnouncementTargeting, error) {
	normalized := AnnouncementTargeting{AnyOf: make([]AnnouncementConditionGroup, 0, len(t.AnyOf))}

	// 允许空 targeting（展示给所有用户）
	if len(t.AnyOf) == 0 {
		return normalized, nil
	}

	if len(t.AnyOf) > 50 {
		return AnnouncementTargeting{}, ErrAnnouncementInvalidTarget
	}

	for _, g := range t.AnyOf {
		if len(g.AllOf) == 0 {
			return AnnouncementTargeting{}, ErrAnnouncementInvalidTarget
		}
		if len(g.AllOf) > 50 {
			return AnnouncementTargeting{}, ErrAnnouncementInvalidTarget
		}

		group := AnnouncementConditionGroup{AllOf: make([]AnnouncementCondition, 0, len(g.AllOf))}
		for _, c := range g.AllOf {
			cond := AnnouncementCondition{
				Type:     strings.TrimSpace(c.Type),
				Operator: strings.TrimSpace(c.Operator),
				Value:    c.Value,
			}
			seenGroupIDs := make(map[string]struct{})
			for _, gid := range c.GroupIDs {
				gid = strings.TrimSpace(gid)
				if gid == "" {
					return AnnouncementTargeting{}, ErrAnnouncementInvalidTarget
				}
				if _, ok := seenGroupIDs[gid]; !ok {
					seenGroupIDs[gid] = struct{}{}
					cond.GroupIDs = append(cond.GroupIDs, gid)
				}
			}
			seenUserIDs := make(map[string]struct{})
			for _, userID := range c.UserIDs {
				userID = strings.TrimSpace(userID)
				if userID == "" {
					return AnnouncementTargeting{}, ErrAnnouncementInvalidTarget
				}
				if _, ok := seenUserIDs[userID]; !ok {
					seenUserIDs[userID] = struct{}{}
					cond.UserIDs = append(cond.UserIDs, userID)
				}
			}
			seenLevels := make(map[int]struct{})
			for _, level := range c.Levels {
				if level < 1 || level > 3 {
					return AnnouncementTargeting{}, ErrAnnouncementInvalidTarget
				}
				if _, ok := seenLevels[level]; !ok {
					seenLevels[level] = struct{}{}
					cond.Levels = append(cond.Levels, level)
				}
			}

			if err := cond.validate(); err != nil {
				return AnnouncementTargeting{}, err
			}
			group.AllOf = append(group.AllOf, cond)
		}

		normalized.AnyOf = append(normalized.AnyOf, group)
	}

	return normalized, nil
}

func (c AnnouncementCondition) validate() error {
	switch c.Type {
	case AnnouncementConditionTypeSubscription:
		if c.Operator != AnnouncementOperatorIn {
			return ErrAnnouncementInvalidTarget
		}
		if len(c.GroupIDs) == 0 {
			return ErrAnnouncementInvalidTarget
		}
		return nil

	case AnnouncementConditionTypeBalance:
		switch c.Operator {
		case AnnouncementOperatorGT, AnnouncementOperatorGTE, AnnouncementOperatorLT, AnnouncementOperatorLTE, AnnouncementOperatorEQ:
			return nil
		default:
			return ErrAnnouncementInvalidTarget
		}

	case AnnouncementConditionTypeUser:
		if c.Operator != AnnouncementOperatorIn || len(c.UserIDs) == 0 {
			return ErrAnnouncementInvalidTarget
		}
		return nil

	case AnnouncementConditionTypeLevel:
		if c.Operator != AnnouncementOperatorIn || len(c.Levels) == 0 {
			return ErrAnnouncementInvalidTarget
		}
		for _, level := range c.Levels {
			if level < 1 || level > 3 {
				return ErrAnnouncementInvalidTarget
			}
		}
		return nil

	default:
		return ErrAnnouncementInvalidTarget
	}
}

type Announcement struct {
	ID         string
	Title      string
	Content    string
	Status     string
	NotifyMode string
	Targeting  AnnouncementTargeting
	StartsAt   *time.Time
	EndsAt     *time.Time
	CreatedBy  *string
	UpdatedBy  *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (a *Announcement) IsActiveAt(now time.Time) bool {
	if a == nil {
		return false
	}
	if a.Status != AnnouncementStatusActive {
		return false
	}
	if a.NotifyMode == AnnouncementNotifyModeSilent {
		return true
	}
	if a.StartsAt != nil && now.Before(*a.StartsAt) {
		return false
	}
	if a.EndsAt != nil && !now.Before(*a.EndsAt) {
		// ends_at 语义：到点即下线
		return false
	}
	return true
}
