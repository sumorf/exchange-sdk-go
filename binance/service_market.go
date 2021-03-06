package binance

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

func (as *apiService) GetAllSymbol() ([]global.TradeSymbol, error) {
	r := &struct {
		Symbols []TradePair `json:"symbols"`
	}{}
	err := as.request("GET", "api/v1/exchangeInfo", nil, &r, false, false)
	if err != nil {
		return nil, err
	}
	rr := []global.TradeSymbol{}
	for _, s := range r.Symbols {
		rr = append(rr, global.TradeSymbol{
			Base:  s.Base,
			Quote: s.Quote,
		})
	}
	return rr, nil
}

// func (as *apiService) SubDepth(sreq global.TradeSymbol) (chan global.Depth, error) {
// 	params := make(map[string]string)
// 	params["symbol"] = strings.ToUpper(sreq.Base + sreq.Quote)
// 	params["limit"] = "100"

// 	ch := make(chan global.Depth, 100)

// 	go func() {
// 		rand.Seed(time.Now().Unix())
// 		rms := rand.Intn(2000)
// 		time.Sleep(time.Duration(rms) * time.Millisecond)
// 		for {
// 			rawBook := &struct {
// 				LastUpdateID int             `json:"lastUpdateId"`
// 				Bids         [][]interface{} `json:"bids"`
// 				Asks         [][]interface{} `json:"asks"`
// 			}{}
// 			err := as.request("GET", "api/v1/depth", params, &rawBook, false, false)
// 			if err != nil {
// 				log.Printf("binance depth error : %+v\n", err)
// 				time.Sleep(time.Duration(rand.Intn(2000)+10000) * time.Millisecond)
// 				continue
// 			}
// 			extractOrder := func(rawPrice, rawQuantity interface{}) (*Order, error) {
// 				price, err := floatFromString(rawPrice)
// 				if err != nil {
// 					return nil, nil
// 				}
// 				quantity, err := floatFromString(rawQuantity)
// 				if err != nil {
// 					return nil, nil
// 				}
// 				return &Order{
// 					Price:    price,
// 					Quantity: quantity,
// 				}, nil
// 			}
// 			r := global.Depth{
// 				Base:  sreq.Base,
// 				Quote: sreq.Quote,
// 				Asks:  []global.DepthPair{},
// 				Bids:  []global.DepthPair{},
// 			}
// 			for _, bid := range rawBook.Bids {
// 				order, err := extractOrder(bid[0], bid[1])
// 				if err != nil {
// 					continue
// 				}
// 				r.Bids = append(r.Bids, global.DepthPair{
// 					Price: order.Price,
// 					Size:  order.Quantity,
// 				})
// 			}
// 			for _, ask := range rawBook.Asks {
// 				order, err := extractOrder(ask[0], ask[1])
// 				if err != nil {
// 					continue
// 				}
// 				r.Asks = append(r.Asks, global.DepthPair{
// 					Price: order.Price,
// 					Size:  order.Quantity,
// 				})
// 			}
// 			// cvt

// 			ch <- r
// 			time.Sleep(time.Duration(rand.Intn(2000)+10000) * time.Millisecond)
// 		}
// 	}()

// 	return ch, nil
// }

func (as *apiService) GetDepth(sreq global.TradeSymbol) (global.Depth, error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(sreq.Base + sreq.Quote)
	params["limit"] = "100"

	rawBook := &struct {
		LastUpdateID int             `json:"lastUpdateId"`
		Bids         [][]interface{} `json:"bids"`
		Asks         [][]interface{} `json:"asks"`
	}{}
	err := as.request("GET", "api/v1/depth", params, &rawBook, false, false)
	if err != nil {
		log.Printf("binance depth error : %+v\n", err)
		return global.Depth{}, err
	}
	extractOrder := func(rawPrice, rawQuantity interface{}) (*Order, error) {
		price, err := floatFromString(rawPrice)
		if err != nil {
			return nil, nil
		}
		quantity, err := floatFromString(rawQuantity)
		if err != nil {
			return nil, nil
		}
		return &Order{
			Price:    price,
			Quantity: quantity,
		}, nil
	}
	r := global.Depth{
		Base:  sreq.Base,
		Quote: sreq.Quote,
		Asks:  []global.DepthPair{},
		Bids:  []global.DepthPair{},
	}
	for _, bid := range rawBook.Bids {
		order, err := extractOrder(bid[0], bid[1])
		if err != nil {
			continue
		}
		r.Bids = append(r.Bids, global.DepthPair{
			Price: order.Price,
			Size:  order.Quantity,
		})
	}
	for _, ask := range rawBook.Asks {
		order, err := extractOrder(ask[0], ask[1])
		if err != nil {
			continue
		}
		r.Asks = append(r.Asks, global.DepthPair{
			Price: order.Price,
			Size:  order.Quantity,
		})
	}
	return r, nil
}

func (as *apiService) AggTrades(atr AggTradesRequest) ([]*AggTrade, error) {
	params := make(map[string]string)
	params["symbol"] = atr.Symbol
	if atr.FromID != 0 {
		params["fromId"] = strconv.FormatInt(atr.FromID, 10)
	}
	if atr.StartTime != 0 {
		params["startTime"] = strconv.FormatInt(atr.StartTime, 10)
	}
	if atr.EndTime != 0 {
		params["endTime"] = strconv.FormatInt(atr.EndTime, 10)
	}
	if atr.Limit != 0 {
		params["limit"] = strconv.Itoa(atr.Limit)
	}
	rawAggTrades := []struct {
		ID             int    `json:"a"`
		Price          string `json:"p"`
		Quantity       string `json:"q"`
		FirstTradeID   int    `json:"f"`
		LastTradeID    int    `json:"l"`
		Timestamp      int64  `json:"T"`
		BuyerMaker     bool   `json:"m"`
		BestPriceMatch bool   `json:"M"`
	}{}
	err := as.request("GET", "api/v1/aggTrades", params, rawAggTrades, false, false)
	if err != nil {
		return nil, err
	}
	aggTrades := []*AggTrade{}
	for _, rawTrade := range rawAggTrades {
		price, err := floatFromString(rawTrade.Price)
		if err != nil {
			return nil, err
		}
		quantity, err := floatFromString(rawTrade.Quantity)
		if err != nil {
			return nil, err
		}
		t := time.Unix(0, rawTrade.Timestamp*int64(time.Millisecond))

		aggTrades = append(aggTrades, &AggTrade{
			ID:             rawTrade.ID,
			Price:          price,
			Quantity:       quantity,
			FirstTradeID:   rawTrade.FirstTradeID,
			LastTradeID:    rawTrade.LastTradeID,
			Timestamp:      t,
			BuyerMaker:     rawTrade.BuyerMaker,
			BestPriceMatch: rawTrade.BestPriceMatch,
		})
	}
	return aggTrades, nil
}

func (as *apiService) GetKline(kr global.KlineReq) ([]global.Kline, error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(kr.Base + kr.Quote)
	params["interval"] = string(kr.Period)
	if kr.Count != 0 {
		params["limit"] = strconv.FormatInt(kr.Count, 10)
	}
	// if kr.Begin != "" {
	// 	params["startTime"] = strconv.FormatInt(kr.StartTime, 10)
	// }
	// if kr.End != "" {
	// 	params["endTime"] = strconv.FormatInt(kr.EndTime, 10)
	// }
	rawKlines := [][]interface{}{}
	err := as.request("GET", "api/v1/klines", params, &rawKlines, false, false)
	if err != nil {
		return nil, err
	}
	klines := []global.Kline{}
	for _, k := range rawKlines {
		ot, err := timeFromUnixTimestampFloat(k[0])
		if err != nil {
			return nil, warpError(err, "cannot parse Kline.OpenTime")
		}
		open, err := floatFromString(k[1])
		if err != nil {
			return nil, warpError(err, "cannot parse Kline.Open")
		}
		high, err := floatFromString(k[2])
		if err != nil {
			return nil, warpError(err, "cannot parse Kline.High")
		}
		low, err := floatFromString(k[3])
		if err != nil {
			return nil, warpError(err, "cannot parse Kline.Low")
		}
		cls, err := floatFromString(k[4])
		if err != nil {
			return nil, warpError(err, "cannot parse Kline.Close")
		}
		volume, err := floatFromString(k[5])
		if err != nil {
			return nil, warpError(err, "cannot parse Kline.Volume")
		}
		// ct, err := timeFromUnixTimestampFloat(k[6])
		// if err != nil {
		// 	return nil, warpError(err, "cannot parse Kline.CloseTime")
		// }
		// qav, err := floatFromString(k[7])
		// if err != nil {
		// 	return nil, warpError(err, "cannot parse Kline.QuoteAssetVolume")
		// }
		// not, ok := k[8].(float64)
		// if !ok {
		// 	return nil, warpError(err, "cannot parse Kline.NumberOfTrades")
		// }
		// tbbav, err := floatFromString(k[9])
		// if err != nil {
		// 	return nil, warpError(err, "cannot parse Kline.TakerBuyBaseAssetVolume")
		// }
		// tbqav, err := floatFromString(k[10])
		// if err != nil {
		// 	return nil, warpError(err, "cannot parse Kline.TakerBuyQuoteAssetVolume")
		// }
		klines = append(klines, global.Kline{
			//OpenTime:                 ot,
			Base:      kr.Base,
			Quote:     kr.Quote,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     cls,
			Volume:    volume,
			Timestamp: ot.Unix() * 1000,
			//CloseTime:                ct,
			//QuoteAssetVolume:         qav,
			//NumberOfTrades:           int(not),
			//TakerBuyBaseAssetVolume:  tbbav,
			//TakerBuyQuoteAssetVolume: tbqav,
		})
	}
	return klines, nil
}

func (as *apiService) Ticker24(symbol string) (*Ticker24, error) {
	params := make(map[string]string)
	params["symbol"] = symbol
	rawTicker24 := struct {
		PriceChange        string  `json:"priceChange"`
		PriceChangePercent string  `json:"priceChangePercent"`
		WeightedAvgPrice   string  `json:"weightedAvgPrice"`
		PrevClosePrice     string  `json:"prevClosePrice"`
		LastPrice          string  `json:"lastPrice"`
		BidPrice           string  `json:"bidPrice"`
		AskPrice           string  `json:"askPrice"`
		OpenPrice          string  `json:"openPrice"`
		HighPrice          string  `json:"highPrice"`
		LowPrice           string  `json:"lowPrice"`
		Volume             string  `json:"volume"`
		OpenTime           float64 `json:"openTime"`
		CloseTime          float64 `json:"closeTime"`
		FirstID            int
		LastID             int
		Count              int
	}{}
	err := as.request("GET", "api/v1/ticker/24hr", params, &rawTicker24, false, false)
	if err != nil {
		return nil, err
	}
	pc, err := strconv.ParseFloat(rawTicker24.PriceChange, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.PriceChange")
	}
	pcPercent, err := strconv.ParseFloat(rawTicker24.PriceChangePercent, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.PriceChangePercent")
	}
	wap, err := strconv.ParseFloat(rawTicker24.WeightedAvgPrice, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.WeightedAvgPrice")
	}
	pcp, err := strconv.ParseFloat(rawTicker24.PrevClosePrice, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.PrevClosePrice")
	}
	lastPrice, err := strconv.ParseFloat(rawTicker24.LastPrice, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.LastPrice")
	}
	bp, err := strconv.ParseFloat(rawTicker24.BidPrice, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.BidPrice")
	}
	ap, err := strconv.ParseFloat(rawTicker24.AskPrice, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.AskPrice")
	}
	op, err := strconv.ParseFloat(rawTicker24.OpenPrice, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.OpenPrice")
	}
	hp, err := strconv.ParseFloat(rawTicker24.HighPrice, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.HighPrice")
	}
	lowPrice, err := strconv.ParseFloat(rawTicker24.LowPrice, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.LowPrice")
	}
	vol, err := strconv.ParseFloat(rawTicker24.Volume, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.Volume")
	}
	ot, err := timeFromUnixTimestampFloat(rawTicker24.OpenTime)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.OpenTime")
	}
	ct, err := timeFromUnixTimestampFloat(rawTicker24.CloseTime)
	if err != nil {
		return nil, warpError(err, "cannot parse Ticker24.CloseTime")
	}
	t24 := &Ticker24{
		PriceChange:        pc,
		PriceChangePercent: pcPercent,
		WeightedAvgPrice:   wap,
		PrevClosePrice:     pcp,
		LastPrice:          lastPrice,
		BidPrice:           bp,
		AskPrice:           ap,
		OpenPrice:          op,
		HighPrice:          hp,
		LowPrice:           lowPrice,
		Volume:             vol,
		OpenTime:           ot,
		CloseTime:          ct,
		FirstID:            rawTicker24.FirstID,
		LastID:             rawTicker24.LastID,
		Count:              rawTicker24.Count,
	}
	return t24, nil
}

func (as *apiService) TickerAllPrices() ([]*PriceTicker, error) {
	params := make(map[string]string)
	rawTickerAllPrices := []struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}{}
	err := as.request("GET", "api/v1/ticker/allPrices", params, rawTickerAllPrices, false, false)
	if err != nil {
		return nil, err
	}

	var tpc []*PriceTicker
	for _, rawTickerPrice := range rawTickerAllPrices {
		p, err := strconv.ParseFloat(rawTickerPrice.Price, 64)
		if err != nil {
			return nil, warpError(err, "cannot parse TickerAllPrices.Price")
		}
		tpc = append(tpc, &PriceTicker{
			Symbol: rawTickerPrice.Symbol,
			Price:  p,
		})
	}
	return tpc, nil
}

func (as *apiService) TickerAllBooks() ([]*BookTicker, error) {
	params := make(map[string]string)
	rawBookTickers := []struct {
		Symbol   string `json:"symbol"`
		BidPrice string `json:"bidPrice"`
		BidQty   string `json:"bidQty"`
		AskPrice string `json:"askPrice"`
		AskQty   string `json:"askQty"`
	}{}
	err := as.request("GET", "api/v1/ticker/allBookTickers", params, rawBookTickers, false, false)
	if err != nil {
		return nil, err
	}
	var btc []*BookTicker
	for _, rawBookTicker := range rawBookTickers {
		bp, err := strconv.ParseFloat(rawBookTicker.BidPrice, 64)
		if err != nil {
			return nil, warpError(err, "cannot parse TickerBookTickers.BidPrice")
		}
		bqty, err := strconv.ParseFloat(rawBookTicker.BidQty, 64)
		if err != nil {
			return nil, warpError(err, "cannot parse TickerBookTickers.BidQty")
		}
		ap, err := strconv.ParseFloat(rawBookTicker.AskPrice, 64)
		if err != nil {
			return nil, warpError(err, "cannot parse TickerBookTickers.AskPrice")
		}
		aqty, err := strconv.ParseFloat(rawBookTicker.AskQty, 64)
		if err != nil {
			return nil, warpError(err, "cannot parse TickerBookTickers.AskQty")
		}
		btc = append(btc, &BookTicker{
			Symbol:   rawBookTicker.Symbol,
			BidPrice: bp,
			BidQty:   bqty,
			AskPrice: ap,
			AskQty:   aqty,
		})
	}
	return btc, nil
}
