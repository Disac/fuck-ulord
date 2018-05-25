package main

import (
	"gopkg.in/redis.v4"
	"log"
	"sort"
	"strconv"
	"strings"
)

const allReward = 112.966

type RedisProvider struct {
	client   *redis.Client
	Addr     string
	Password string
	PoolSize int
}

func (rp *RedisProvider) Init() {
	rp.client = redis.NewClient(&redis.Options{
		Addr:     rp.Addr,
		Password: rp.Password,
		PoolSize: rp.PoolSize,
	})
}
func (rp *RedisProvider) Provide() interface{} {
	return rp.client
}

func main() {
	re := &RedisProvider{
		Addr:     "r-m5e83220c565ce74.redis.rds.aliyuncs.com:6379",
		Password: "Mengxiaozhu123",
		PoolSize: 100,
	}
	re.Init()
	log.Println(re.Password, re.Addr, re.PoolSize)
	cli := re.Provide().(*redis.Client)

	Go(cli, "ulord:shares:roundCurrent", "ulord:shares:timesCurrent")
	Go(cli, "ulord{:shares:round}Current", "ulord{:shares:times}Current")
}

func Times(round, shares map[string]string) map[string]float64 {
	timesToAddr := make(map[string]float64)
	sharesToAddr := make(map[string]float64)

	sum := 0.0
	for k, v := range shares {
		address := strings.Split(k, ".")[0]
		val, _ := strconv.ParseFloat(v, 64)
		_, has := sharesToAddr[address]
		if !has {
			sharesToAddr[address] = val
		}
		sum += sharesToAddr[address]
	}

	var times []float64
	for k, v := range round {
		val, _ := strconv.ParseFloat(v, 64)
		times = append(times, val)
		address := strings.Split(k, ".")[0]
		last, has := timesToAddr[address]
		if !has {
			timesToAddr[address] = val
			continue
		}

		if last >= val {
			timesToAddr[address] += val / 2
		} else {
			timesToAddr[address] = val + last/2
		}
	}

	sort.Float64s(times)
	maxTimes := times[len(times)-1]

	for addr, t := range timesToAddr {
		if t < maxTimes*0.51 {
			lostShare := sharesToAddr[addr] * (1 - t/maxTimes)
			sharesToAddr[addr] = sharesToAddr[addr] - lostShare
		}
	}

	reward := make(map[string]float64)

	for addr, share := range sharesToAddr {
		percent := share / sum
		reward[addr] = allReward * percent
	}
	return reward
}

func Go(cli *redis.Client, sharekey, timeskey string) {
	roundCurrent := cli.HGetAll(sharekey)
	round, err := roundCurrent.Result()
	if err != nil {
		log.Println(err)
		return
	}
	timesCurrent := cli.HGetAll(timeskey)
	times, err := timesCurrent.Result()
	if err != nil {
		log.Println(err)
		return
	}
	reward := Times(round, times)

	for k, v := range reward {
		log.Println(k, v)
	}
}
