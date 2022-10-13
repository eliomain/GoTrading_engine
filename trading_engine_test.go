package trading_engine

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
)

var btcusdt = NewTradePair("btcusdt", 2, 0)

func d(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

func TestTradePairDepth(t *testing.T) {

	Convey("卖盘深度行情", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(1.01), d(2), 1112))
		btcusdt.PushNewOrder(NewAskLimitItem("id2", d(1.01), d(2), 1113))
		btcusdt.PushNewOrder(NewAskLimitItem("id3", d(1.1), d(2), 1114))
		time.Sleep(time.Millisecond * time.Duration(100))
		ask := btcusdt.GetAskDepth(0)
		str_ask, _ := json.Marshal(ask)
		So(string(str_ask), ShouldEqual, `[["1.01","4"],["1.10","2"]]`)
	})

	Convey("买盘深度行情", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewBidLimitItem("id4", d(1.02), d(2), 1115))
		btcusdt.PushNewOrder(NewBidLimitItem("id5", d(1.3), d(2), 1116))
		btcusdt.PushNewOrder(NewBidLimitItem("id6", d(1.02), d(2), 1117))
		btcusdt.PushNewOrder(NewBidLimitItem("id7", d(0.02), d(1), 1118))

		time.Sleep(time.Millisecond * time.Duration(100))
		bid := btcusdt.GetBidDepth(0)
		str_bid, _ := json.Marshal(bid)
		So(string(str_bid), ShouldEqual, `[["1.30","2"],["1.02","4"],["0.02","1"]]`)
	})
}

func TestTradeFunc_LimitOrder(t *testing.T) {
	//创建一个买单
	Convey("新增一个限价买单", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewBidLimitItem("id1", d(1.1), d(1.2), 1112))

		So(btcusdt.askQueue.Len(), ShouldEqual, 0)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 1)
		So(btcusdt.bidQueue.Top().GetPrice(), ShouldEqual, d(1.1))
		So(btcusdt.bidQueue.Top().GetUniqueId(), ShouldEqual, "id1")
		So(btcusdt.bidQueue.Top().GetQuantity(), ShouldEqual, d(1.2))
		So(btcusdt.bidQueue.Top().GetCreateTime(), ShouldEqual, 1112)
	})

	Convey("新增一个限价卖单", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(1.1), d(1.2), 1112))

		So(btcusdt.askQueue.Len(), ShouldEqual, 1)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 0)
		So(btcusdt.askQueue.Top().GetPrice(), ShouldEqual, d(1.1))
		So(btcusdt.askQueue.Top().GetUniqueId(), ShouldEqual, "id1")
		So(btcusdt.askQueue.Top().GetQuantity(), ShouldEqual, d(1.2))
		So(btcusdt.askQueue.Top().GetCreateTime(), ShouldEqual, 1112)

	})

	Convey("清空买卖队列", t, func() {
		btcusdt.cleanAll()
		So(btcusdt.askQueue.Len(), ShouldEqual, 0)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 0)
	})

	Convey("限价买卖单，价格一致，完全成交", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(1.1), d(1.2), 1112))
		btcusdt.PushNewOrder(NewBidLimitItem("id2", d(1.1), d(1.2), 1113))
		tradeLog := <-btcusdt.ChTradeResult
		So(tradeLog.AskOrderId, ShouldEqual, "id1")
		So(tradeLog.BidOrderId, ShouldEqual, "id2")
		So(tradeLog.TradePrice, ShouldEqual, d(1.1))
		So(tradeLog.TradeQuantity, ShouldEqual, d(1.2))
	})

	Convey("限价买卖单，价格一致，买单部分成交", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(1.1), d(1.2), 1112))
		btcusdt.PushNewOrder(NewBidLimitItem("id2", d(1.1), d(2.3), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id1")
		So(tradeLog.BidOrderId, ShouldEqual, "id2")
		So(tradeLog.TradePrice, ShouldEqual, d(1.1))
		So(tradeLog.TradeQuantity, ShouldEqual, d(1.2))
		So(btcusdt.bidQueue.Top().GetQuantity(), ShouldEqual, d(1.1))
		So(btcusdt.bidQueue.Len(), ShouldEqual, 1)
		So(btcusdt.askQueue.Len(), ShouldEqual, 0)
	})

	Convey("限价买卖单，价格一致，卖单部分成交", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(1.1), d(2.2), 1112))
		btcusdt.PushNewOrder(NewBidLimitItem("id2", d(1.1), d(1.3), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id1")
		So(tradeLog.BidOrderId, ShouldEqual, "id2")
		So(tradeLog.TradePrice, ShouldEqual, d(1.1))
		So(tradeLog.TradeQuantity, ShouldEqual, d(1.3))
		So(btcusdt.askQueue.Top().GetQuantity(), ShouldEqual, d(0.9))
		So(btcusdt.askQueue.Len(), ShouldEqual, 1)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 0)
	})

	Convey("价格一致，时间优先", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(1.1), d(2.2), 1112))
		btcusdt.PushNewOrder(NewAskLimitItem("id2", d(1.1), d(2.2), 1110))

		btcusdt.PushNewOrder(NewBidLimitItem("id3", d(1.1), d(1.3), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id2")
		So(tradeLog.BidOrderId, ShouldEqual, "id3")
		So(tradeLog.TradePrice, ShouldEqual, d(1.1))
		So(tradeLog.TradeQuantity, ShouldEqual, d(1.3))
		So(btcusdt.askQueue.Top().GetQuantity(), ShouldEqual, d(0.9))
		So(btcusdt.askQueue.Len(), ShouldEqual, 2)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 0)
	})

	//价格优先
	Convey("价格优先", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(1.01), d(2.2), 1112))
		btcusdt.PushNewOrder(NewAskLimitItem("id2", d(1.1), d(2.2), 1110))

		btcusdt.PushNewOrder(NewBidLimitItem("id3", d(1.1), d(1.3), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id1")
		So(tradeLog.BidOrderId, ShouldEqual, "id3")
		So(tradeLog.TradePrice, ShouldEqual, d(1.01))
		So(tradeLog.TradeQuantity, ShouldEqual, d(1.3))
		So(btcusdt.askQueue.Top().GetQuantity(), ShouldEqual, d(0.9))
		So(btcusdt.askQueue.Len(), ShouldEqual, 2)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 0)
	})

}

func TestTradeFunc_MarketBuyOrder(t *testing.T) {

	Convey("市价买入 按数量买入, 金额足够买单完全成交", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(1.01), d(2.2), 1112))
		btcusdt.PushNewOrder(NewBidMarketQtyItem("id2", d(1.1), d(100), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id1")
		So(tradeLog.BidOrderId, ShouldEqual, "id2")
		So(tradeLog.TradePrice, ShouldEqual, d(1.01))
		So(tradeLog.TradeQuantity, ShouldEqual, d(1.1))
		So(tradeLog.MarketOrder, ShouldEqual, "id2")

	})

	Convey("市价按数量买入, 账户金额足够买单部分成交", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(1.01), d(2.2), 1112))
		btcusdt.PushNewOrder(NewBidMarketQtyItem("id2", d(100), d(100), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id1")
		So(tradeLog.BidOrderId, ShouldEqual, "id2")
		So(tradeLog.TradePrice, ShouldEqual, d(1.01))
		So(tradeLog.TradeQuantity, ShouldEqual, d(2.2))
		So(tradeLog.MarketOrder, ShouldEqual, "id2")
	})

	//市价买入 按数量, 金额不足 买单部分成交
	Convey("市价按数量买入, 账户金额不足, 买单部分成交", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(100), d(20), 1112))
		btcusdt.PushNewOrder(NewBidMarketQtyItem("id2", d(20), d(100), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id1")
		So(tradeLog.BidOrderId, ShouldEqual, "id2")
		So(tradeLog.TradePrice, ShouldEqual, d(100))
		So(tradeLog.TradeQuantity, ShouldEqual, d(1))

		So(tradeLog.MarketOrder, ShouldEqual, "id2")
		So(btcusdt.askQueue.Len(), ShouldEqual, 1)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 0)
	})

	Convey("市价买入 指定金额, 买单完全成交", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(10.00), d(100), 1112))
		btcusdt.PushNewOrder(NewBidMarketAmountItem("id2", d(50), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id1")
		So(tradeLog.BidOrderId, ShouldEqual, "id2")
		So(tradeLog.TradePrice, ShouldEqual, d(10.00))
		So(tradeLog.TradeQuantity, ShouldEqual, d(5))

		So(tradeLog.MarketOrder, ShouldEqual, "id2")
		So(btcusdt.askQueue.Len(), ShouldEqual, 1)
	})

	Convey("市价买入 指定金额, 买单部分成交", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewAskLimitItem("id1", d(10.00), d(100), 1112))
		btcusdt.PushNewOrder(NewBidMarketAmountItem("id2", d(6000), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id1")
		So(tradeLog.BidOrderId, ShouldEqual, "id2")
		So(tradeLog.TradePrice, ShouldEqual, d(10.00))
		So(tradeLog.TradeQuantity, ShouldEqual, d(100))

		So(tradeLog.MarketOrder, ShouldEqual, "id2")
		So(btcusdt.askQueue.Len(), ShouldEqual, 0)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 0)
	})
}

func TestTradeFunc_MarketSellOrder(t *testing.T) {
	Convey("市价卖出 按数量, 完全成交", t, func() {

		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewBidLimitItem("id1", d(10.00), d(100), 1112))
		btcusdt.PushNewOrder(NewAskMarketQtyItem("id2", d(6), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id2")
		So(tradeLog.BidOrderId, ShouldEqual, "id1")
		So(tradeLog.TradePrice, ShouldEqual, d(10.00))
		So(tradeLog.TradeQuantity, ShouldEqual, d(6))

		So(tradeLog.MarketOrder, ShouldEqual, "id2")
		So(btcusdt.askQueue.Len(), ShouldEqual, 0)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 1)
	})

	Convey("市价卖出 按数量，部分成交", t, func() {
		//
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewBidLimitItem("id1", d(10.00), d(100), 1112))
		btcusdt.PushNewOrder(NewAskMarketQtyItem("id2", d(6000), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id2")
		So(tradeLog.BidOrderId, ShouldEqual, "id1")
		So(tradeLog.TradePrice, ShouldEqual, d(10.00))
		So(tradeLog.TradeQuantity, ShouldEqual, d(100))

		So(tradeLog.MarketOrder, ShouldEqual, "id2")
		So(btcusdt.askQueue.Len(), ShouldEqual, 0)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 0)
	})

	Convey("市价卖出 指定金额，持仓足够完全成交", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewBidLimitItem("id1", d(10.00), d(1000), 1112))
		btcusdt.PushNewOrder(NewAskMarketAmountItem("id2", d(6000), d(1000000), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id2")
		So(tradeLog.BidOrderId, ShouldEqual, "id1")
		So(tradeLog.TradePrice, ShouldEqual, d(10.00))
		So(tradeLog.TradeQuantity, ShouldEqual, d(600))

		So(tradeLog.MarketOrder, ShouldEqual, "id2")
		So(btcusdt.askQueue.Len(), ShouldEqual, 0)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 1)
	})

	Convey("市价卖出 指定金额，持仓足够 但是部分成交", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewBidLimitItem("id1", d(10.00), d(50), 1112))
		btcusdt.PushNewOrder(NewAskMarketAmountItem("id2", d(6000), d(1000000), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id2")
		So(tradeLog.BidOrderId, ShouldEqual, "id1")
		So(tradeLog.TradePrice, ShouldEqual, d(10.00))
		So(tradeLog.TradeQuantity, ShouldEqual, d(50))

		So(tradeLog.MarketOrder, ShouldEqual, "id2")
		So(btcusdt.askQueue.Len(), ShouldEqual, 0)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 0)
	})

	Convey("市价卖出 指定金额，持仓不足", t, func() {
		btcusdt.cleanAll()
		btcusdt.PushNewOrder(NewBidLimitItem("id1", d(100.00), d(50), 1112))
		btcusdt.PushNewOrder(NewAskMarketAmountItem("id2", d(500), d(3), 1113))
		tradeLog := <-btcusdt.ChTradeResult

		So(tradeLog.AskOrderId, ShouldEqual, "id2")
		So(tradeLog.BidOrderId, ShouldEqual, "id1")
		So(tradeLog.TradePrice, ShouldEqual, d(100))
		So(tradeLog.TradeQuantity, ShouldEqual, d(3))

		So(tradeLog.MarketOrder, ShouldEqual, "id2")
		So(btcusdt.askQueue.Len(), ShouldEqual, 0)
		So(btcusdt.bidQueue.Len(), ShouldEqual, 1)
	})
}
