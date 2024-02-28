// SPDX-License-Identifier: MIT

package jwt

type (
	// Blocker 判断令牌是否被丢弃
	//
	// 在某些情况下，需要强制用户的令牌不再可用，可以使用 Blocker 接口。
	Blocker[T Claims] interface {
		// TokenIsBlocked 令牌是否已被提早丢弃
		TokenIsBlocked(string) bool

		// ClaimsIsBlocked 根据 Claims 判断是否已经丢弃
		//
		// 这是对令牌解码之后的阻断行为，性能上有解码的开销，便是相对来说也更加的灵活，
		// 比如要禁用某一用户所有签发的令牌，或是为某一设备签发的令牌等，
		// 只要 T 类型中带的字段均可作为判断依据。
		ClaimsIsBlocked(T) bool
	}

	defaultBlocker[T Claims] struct{}
)

func (d defaultBlocker[T]) TokenIsBlocked(_ string) bool { return false }

func (d defaultBlocker[T]) ClaimsIsBlocked(_ T) bool { return false }
