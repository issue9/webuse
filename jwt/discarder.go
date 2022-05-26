// SPDX-License-Identifier: MIT

package jwt

type (
	// Discarder 判断令牌是否被丢弃
	//
	// 在某些情况下，需要强制用户的令牌不再可用，可以使用 Discarder 接口，
	// 当 JWT 接受此对象时，将采用 IsDiscarded 来判断令牌是否是被丢弃的。
	Discarder[T Claims] interface {
		// TokenIsDiscarded 令牌是否已被提早丢弃
		TokenIsDiscarded(string) bool

		// ClaimsIsDiscarded 根据 Claims 判断是否已经丢弃
		//
		// 这是对令牌解码之后的阻断行为，性能上有解码的开销，便是相对来说也更加的灵活，
		// 比如要禁用某一用户所有签发的令牌，或是为某一设备签发的令牌等，
		// 只要 T 类型中带的字段均可作为判断依据。
		ClaimsIsDiscarded(T) bool
	}

	defaultDiscarder[T Claims] struct{}
)

func (d defaultDiscarder[T]) TokenIsDiscarded(_ string) bool { return false }

func (d defaultDiscarder[T]) ClaimsIsDiscarded(_ T) bool { return false }
