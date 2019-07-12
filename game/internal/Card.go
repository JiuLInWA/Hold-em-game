package internal

// 牌型
type CardSuit int32

const (
	NULL            CardSuit = 0
	HIGH_CARD       CardSuit = 1  // 高牌
	ONE_PAIR        CardSuit = 2  // 一对
	TWO_PAIRS       CardSuit = 3  // 两对
	THREE_OF_A_KIND CardSuit = 4  // 三条
	STRAIGHT        CardSuit = 5  // 顺子
	FLUSH           CardSuit = 6  // 同花
	FULL_HOUSE      CardSuit = 7  // 葫芦
	FOUR_OF_A_KIND  CardSuit = 8  // 四条
	STRAIGHT_FLUSH  CardSuit = 9  // 同花顺
	ROYAL_FLUSH     CardSuit = 10 // 皇家同花顺
)

