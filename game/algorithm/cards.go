package algorithm

type Cards []byte

// 扑克牌52张，分别包含普通牌52张 2-10、J、Q、K、A （以上每种牌4个花色 黑桃、梅花、红心、方块）
var CARDS = []byte{
	//方块
	0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E,
	//梅花
	0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E,
	//红桃
	0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E,
	//黑桃
	0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, 0x3E,
}

const (
	CARDRANK = 13
	SUITSIZE = 4
	TOTAL    = 13 * 4

	TYPE_LEN  = 5 // 牌型长度
	CARDS_LEN = 7 // 河牌+底牌长度
)

//牌型大小：
//1、皇家同花顺＞同花顺＞四条＞葫芦＞同花＞顺子＞三条＞两队＞一对＞单牌
//2、牌点从大到小为：A、K、Q、J、10、9、8、7、6、5、4、3、2，各花色不分大小。
//3、同种牌型，对子时比对子的大小，其它牌型比最大的牌张，如最大牌张相同则比第二大的牌张，以此类推，都相同时为相同。
//4、A 可以组顺子 A、1、2、3、4、5  也可以组顺子 10、J、Q、K、A  同花顺同理
//所有牌型如下：
const (
	ROYAL_FLUSH    uint8= 10  // 皇家同花顺：同一花色的最大顺子。（最大牌：A-K-Q-J-10）
	STRAIGHT_FLUSH uint8= 9  // 同花顺：同一花色的顺子。（最大牌：K-Q-J-10-9 最小牌：A-2-3-4-5）
	FOUR           uint8= 8 // 四条：四同张加单张。（最大牌：A-A-A-A-K 最小牌：2-2-2-2-3）
	FULL_HOUSE     uint8= 7  // 葫芦（豪斯）：三同张加对子。（最大牌：A-A-A-K-K 最小牌：2-2-2-3-3）
	FLUSH          uint8= 6  // 同花：同一花色。（最大牌：A-K-Q-J-9 最小牌：2-3-4-5-7）
	STRAIGHT       uint8= 5 // 顺子：花色不一样的顺子。（最大牌：A-K-Q-J-10 最小牌：A-2-3-4-5）
	THREE          uint8= 4  // 三条：三同张加两单张。（最大牌：A-A-A-K-Q 最小牌：2-2-2-3-4）
	TWO_PAIR       uint8= 3  // 两对：（最大牌：A-A-K-K-Q 最小牌：2-2-3-3-4）
	ONE_PAIR       uint8= 2  // 一对：（最大牌：A-A-K-Q-J 最小牌：2-2-3-4-5）
	HIGH_CARD      uint8= 1 // 高牌：（最大牌：A-K-Q-J-9 最小牌：2-3-4-5-7）
)


// todo 两对和四张起脚牌的判定

var StraightMask = []uint16{15872, 7936, 3968, 1984, 992, 496, 248, 124, 62, 31}
//顺子（Straight，亦称“蛇”）
//此牌由五张顺序扑克牌组成。
//平手牌：如果不止一人抓到此牌，则五张牌中点数最大的赢得此局，
// 如果所有牌点数都相同，平分彩池。
func (this *Cards) straight() uint32 {
	var handvalue uint16
	for _, v := range (*this) {
		value := v & 0xF
		if value == 0xE {
			handvalue |= 1
		}
		handvalue |= (1 << (value - 1 ) )
	}

	for i := uint8(0); i < 10; i++ {
		if handvalue&StraightMask[i] == StraightMask[i] {
			return En(STRAIGHT, uint32(10-i+4))
		}
	}
	return 0
}

//同花顺（Straight Flush）
//五张同花色的连续牌。
//平手牌：如果摊牌时有两副或多副同花顺，连续牌的头张牌大的获得筹码。
//如果是两副或多副相同的连续牌，平分筹码。
func (this *Cards) straightFlush() uint32 {
	cards := *this
	for i := byte(0); i < SUITSIZE; i++ {
		var handvalue uint16
		for _, v := range cards {
			if (v >> 4) == i {
				value := v & 0xF
				if value == 0xE {
					handvalue |= 1
				}
				handvalue |= (1 << (value - 1 ) )
			}
		}

		for i := uint8(0); i < 10; i++ {
			if handvalue&StraightMask[i] == StraightMask[i] {
				return En(STRAIGHT_FLUSH, uint32(10-i+4))
			}
		}
	}
	return 0
}

//皇家同花顺（Royal Flush）
//同花色的A, K, Q, J和10。
//平手牌：在摊牌的时候有两副多副皇家同花顺时，平分筹码。
func (this *Cards) royalFlush() uint32 {
	cards := *this
	for i := byte(0); i < SUITSIZE; i++ {
		var handvalue uint16
		for _, v := range cards {
			if (v >> 4) == i {
				value := v & 0xF
				if value == 0xE {
					handvalue |= 1
				}
				handvalue |= (1 << (value - 1 ) )

			}
		}

		if handvalue&StraightMask[0] == StraightMask[0] {
			return En(ROYAL_FLUSH, 0)
		}
	}
	return 0
}

//四条（Four of a Kind，亦称“铁支”、“四张”或“炸弹”）
//其中四张是相同点数但不同花的扑克牌，第五张是随意的一张牌。
//平手牌：如果两组或者更多组摊牌，则四张牌中的最大者赢局，如果一组人持有的四张牌是一样的，
//那么第五张牌最大者赢局（起脚牌,2张起手牌中小的那张就叫做起脚牌）。如果起脚牌也一样，平分彩池。
func (this *Cards) four(counter *ValueCounter) uint32 {
	cards := *this
	if counter.Get(cards[len(cards)-1]) == 4 {
		return En(FOUR, ToValue(cards))
	}
	return 0
}

//满堂彩（Fullhouse，葫芦，三带二）
//由三张相同点数及任何两张其他相同点数的扑克牌组成。
//平手牌：如果两组或者更多组摊牌，那么三张相同点数中较大者赢局。
//如果三张牌都一样，则两张牌中点数较大者赢局，如果所有的牌都一样，则平分彩池。
func (this *Cards) fullFouse(counter *ValueCounter) uint32 {
	cards := *this
	length := len(cards)

	if length >= 5 {
		if cards[length-1]&0xF == cards[length-2]&0xF &&
			cards[length-3]&0xF == cards[length-1]&0xF &&
			cards[length-4]&0xF == cards[length-5]&0xF {

			return En(FULL_HOUSE, ToValue(cards))
		}
	}
	return 0
}

//同花（Flush，简称“花”）
//此牌由五张不按顺序但相同花的扑克牌组成。
//平手牌：如果不止一人抓到此牌相，则牌点最高的人赢得该局，
//如果最大点相同，则由第二、第三、第四或者第五张牌来决定胜负，如果所有的牌都相同，平分彩池。
func (this *Cards) flush() uint32 {
	cards := *this
	for i := byte(0); i < SUITSIZE; i++ {
		var count uint8
		for _, v := range cards {
			if (v >> 4) == i {
				count ++
				if count == 5 {
					var handvalue uint16
					for _, v1 := range cards {
						if (v1 >> 4) == i {
							value := v1 & 0xF
							if value == 0xE {
								handvalue |= 1
							}
							handvalue |= (1 << (value - 1 ) )
						}
					}
					return En(FLUSH, uint32(handvalue))
				}
			}
		}

	}
	return 0
}

//三条（Three of a kind，亦称“三张”）
//由三张相同点数和两张不同点数的扑克组成。
//平手牌：如果不止一人抓到此牌，则三张牌中最大点数者赢局，
//如果三张牌都相同，比较第四张牌，必要时比较第五张，点数大的人赢局。如果所有牌都相同，则平分彩池。
func (this *Cards) three(counter *ValueCounter) uint32 {
	cards := *this
	if counter.Get(cards[len(cards)-1]) == 3 {
		return En(THREE, ToValue(cards))
	}
	return 0
}

//两对（Two Pairs）
//两对点数相同但两两不同的扑克和随意的一张牌组成。
//平手牌：如果不止一人抓大此牌相，牌点比较大的人赢，如果比较大的牌点相同，那么较小牌点中的较大者赢，
//如果两对牌点相同，那么第五张牌点较大者赢（起脚牌,2张起手牌中小的那张就叫做起脚牌）。如果起脚牌也相同，则平分彩池。
func (this *Cards) twoPair() uint32 {
	cards := *this
	length := len(cards)
	if length >= 4 {
		if cards[length-1]&0xF == cards[length-2]&0xF &&
			cards[length-3]&0xF == cards[length-4]&0xF {
			return En(TWO_PAIR, ToValue(cards))
		}
	}
	return 0
}

//一对（One Pair）
//由两张相同点数的扑克牌和另三张随意的牌组成。
//平手牌：如果不止一人抓到此牌，则两张牌中点数大的赢，如果对牌都一样，则比较另外三张牌中大的赢，
//如果另外三张牌中较大的也一样则比较第二大的和第三大的，如果所有的牌都一样，则平分彩池。
func (this *Cards) onePair() uint32 {
	cards := *this
	length := len(cards)
	if length >= 2 {
		if cards[length-1]&0xF == cards[length-2]&0xF {
			return En(ONE_PAIR, ToValue(cards))
		}
	}
	return 0
}

func ToValue(cards []byte) uint32 {
	var res uint32
	for i := len(cards) - 1; i >= 0; i-- {
		res *= 10
		res += uint32(cards[i] & 0xF)
	}
	return res
}

func De(v uint32) (uint8, uint32) {
	return uint8(v >> 24), v & 0xFFFFFF
}

func En(t uint8, v uint32) uint32 {
	v1 := v | ( uint32(t) << 24)
	return v1
}
