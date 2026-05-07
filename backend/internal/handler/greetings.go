package handler

import (
	"math/rand"
	"time"
)

// greetingTemplates holds 50 warm, human-sounding greeting templates.
// Organized by time-gap categories for contextual relevance.

var greetingCategories = []struct {
	maxGap   time.Duration
	messages []string
}{
	// 类别 0: 首次使用 (无历史消息)
	{maxGap: 0, messages: []string{
		"嗨，我是年糕。有空的时候可以把你收藏的经验讲给我听，我们慢慢聊。",
		"你好呀。这里没有固定的开场白——你想从哪儿说起？",
		"嘿，来了。你可以和我聊聊最近在想的事，也可以翻翻你收藏的经验，我们一起来看。",
		"你好，我是年糕。想聊什么都可以，不用想太多，随便说说。",
		"嗨。你收藏的经验我都记得，想聊哪条随时告诉我。不急，慢慢来。",
	}},
	// 类别 1: 2-12 小时
	{maxGap: 12 * time.Hour, messages: []string{
		"回来啦。今天过得怎么样？",
		"刚才在想什么？",
		"又见面了。刚才那段忙完了？",
		"嗨。今天的节奏还好吗？",
		"嗯，你来了。有什么想说的吗？",
		"今天心情怎么样？随便聊聊。",
		"刚才处理的事情顺利吗？",
		"来了。状态还行吗？",
		"嘿。下午的时光怎么过的？",
		"又到了可以停下来聊聊的时间了。",
	}},
	// 类别 2: 12-24 小时
	{maxGap: 24 * time.Hour, messages: []string{
		"昨天休息得还好吗。有什么想聊的随时说。",
		"新的一天。今天有什么打算？",
		"嗨，早上好（也许是晚上）。这几天在琢磨什么？",
		"昨天的状态怎么样？有发生什么想聊聊的事吗？",
		"休息好了吗。慢慢来，我一直在。",
		"一天过去了。有什么想复盘或者分享的吗？",
		"嗯，你来了。今天有没有遇到让你在意的事？",
		"昨天睡得好不好？今天可以随便聊聊。",
		"又过了一天。有什么新的想法吗？",
		"嗨。不管今天过得怎样，这里都可以聊聊。",
	}},
	// 类别 3: 1-3 天
	{maxGap: 72 * time.Hour, messages: []string{
		"这两天忙什么了？我一直在。",
		"有一阵子没聊了。最近有什么新的感悟吗？",
		"嘿。这几天过得怎么样？",
		"你回来了。这几天有发生什么事想聊聊的吗？",
		"几天没见。有什么变化吗？",
		"嗨。这几天在忙什么？",
		"这几天还顺利吗？不急，慢慢说。",
		"好久没聊了。还是老样子，你想从哪儿开始？",
		"嗯，你来了。这几天有没有想到过之前的哪条经验？",
		"几天过去了。想聊聊这段时间的收获或者困惑吗？",
	}},
	// 类别 4: 3-7 天
	{maxGap: 7 * 24 * time.Hour, messages: []string{
		"有一周没见了。最近有什么新的经历吗？",
		"嘿，好久没聊。这一周还好吗？",
		"你来了。一周可以发生很多事，想聊聊什么？",
		"一周过去了。有什么想说的吗？",
		"好久不见。最近有没有收藏什么新的经验想聊聊？",
		"嗯。这段时间还顺利吗？",
		"又见面了。这一周有没有什么让你印象深刻的？",
		"嗨。不管过了多久，你想聊的时候我都在。",
	}},
	// 类别 5: 7-30 天
	{maxGap: 30 * 24 * time.Hour, messages: []string{
		"好久不见。这段时间发生了什么？慢慢聊。",
		"你回来了。这段时间肯定有不少新的经历吧。",
		"好久不见。想聊聊最近的事吗？",
		"有一阵子没见了。最近还好吗？",
		"嘿，欢迎回来。过了这么久，想从哪说起？",
		"好久没和你聊天了。这段时间有什么想分享的吗？",
		"你终于来了。我一直记得你收藏的经验，想聊哪条？",
	}},
}

// pickGreeting selects a greeting based on time since last message.
// It picks the correct time-gap category and returns a random message from that category.
func pickGreeting(lastMessageAt time.Time) string {
	gap := time.Since(lastMessageAt)
	return pickGreetingByGap(gap)
}

// pickWelcomeGreeting returns a greeting for first-time users.
func pickWelcomeGreeting() string {
	msgs := greetingCategories[0].messages
	return msgs[rand.Intn(len(msgs))]
}

// pickGreetingByGap selects a greeting based on time gap.
func pickGreetingByGap(gap time.Duration) string {
	// Find the right category
	var pool []string
	for i := len(greetingCategories) - 1; i >= 0; i-- {
		if greetingCategories[i].maxGap == 0 {
			continue // skip "first time" category
		}
		if gap <= greetingCategories[i].maxGap || i == len(greetingCategories)-1 {
			pool = greetingCategories[i].messages
			break
		}
	}
	if pool == nil {
		// Fallback: last category
		pool = greetingCategories[len(greetingCategories)-1].messages
	}
	return pool[rand.Intn(len(pool))]
}
