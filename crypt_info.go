package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"sync"
	"github.com/kataras/iris/websocket"
)

type (
	Currency struct {
		Id 						string					`json:"id"`
		Name 					string					`json:"name"`
		Symbol 					string					`json:"symbol"`
		Rank 					string					`json:"rank"`
		Price_usd 				string					`json:"price_usd"`
		Price_btc 				string					`json:"price_btc"`
		VolumeUsd24h 			string					`json:"24h_volume_usd"`
		MarketCapUsd 			string					`json:"market_cap_usd"`
		AvailableSupply			string					`json:"available_supply"`
		TotalSupply 			string					`json:"total_supply"`
		MaxSupply 				string					`json:"max_supply"`
		PercentChange1h 		string					`json:"percent_change_1h"`
		PercentChange24h 		string					`json:"percent_change_24h"`
		PercentChange7d 		string					`json:"percent_change_7d"`
		LastUpdated 			string					`json:"last_updated"`
		PriceVnd 				string					`json:"price_vnd"`
		VolumeVnd24h 			string					`json:"24h_volume_vnd"`
		MarketCapVnd 			string					`json:"market_cap_vnd"`
	}
	CurrencyCtrl struct {
		currencies		[]Currency
		mux sync.Mutex
	}
	UserConnects struct {
		mux sync.Mutex
		connects map[websocket.Connection]bool
	}
)

var (
	currencies = CurrencyCtrl{}
	connector = UserConnects{}
)

func initGetData()  {
	connector.Init()
	ticker := time.NewTicker(5 * time.Second)
	getData()
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <- ticker.C:
				getData()
			case <- quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func getData()  {
	resp,err := http.Get("https://api.coinmarketcap.com/v1/ticker/?convert=VND&limit=100")
	if err != nil {
		//log.Panic(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Panic(err)
		return
	}
	cs,err := currencies.Change(body)
	if err != nil {
		//log.Panic(err)
		return
	}
	connector.Send(cs)
}

func (c *CurrencyCtrl) Change(data []byte) ([]Currency,error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	err := json.Unmarshal(data,&c.currencies)
	return c.currencies,err
}

func (c CurrencyCtrl) Get() ([]Currency) {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.currencies
}

func (u *UserConnects) Init() {
	u.connects = make(map[websocket.Connection]bool)
}

func (u *UserConnects) Add(conn websocket.Connection) {
	u.mux.Lock()
	u.connects[conn] = true
	u.mux.Unlock()
}

func (u *UserConnects) Delete(conn websocket.Connection)  {
	u.mux.Lock()
	delete(u.connects, conn)
	u.mux.Unlock()
}

func (u *UserConnects) Send(data []Currency) {
	u.mux.Lock()
	for conn := range u.connects {
		conn.Emit("data",data)
	}
	u.mux.Unlock()
}