package bot

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"devzat/pkg/colors"
	i "devzat/pkg/interfaces"
)

const (
	defaultName = "DevBot"
)

type DevBot struct {
	room  i.Room
	name  string
	rules []responseRule
	colors.Formatter
	colors struct{ fg, bg string }
}

func (b *DevBot) Room() i.Room {
	return b.room
}

func (b *DevBot) SetRoom(r i.Room) {
	b.room = r
}

func (b *DevBot) Server() i.Server {
	return b.room.Server()
}

func (b *DevBot) SetServer(server i.Server) {
}

func (b *DevBot) ChangeColor(colorName string) error {
	//TODO implement me
	panic("implement me")
}

func (b *DevBot) SetForegroundColor(s string) error {
	//TODO implement me
	panic("implement me")
}

func (b *DevBot) SetBackgroundColor(s string) error {
	//TODO implement me
	panic("implement me")
}

func (b *DevBot) Init() error {
	room := b.Room()
	if room == nil {
		return errors.New("this bot doesn't have a room set, yet")
	}

	b.name = b.Colors.Green.Paint(defaultName)

	return nil
}

func (b *DevBot) Name() string {
	return b.name
}

func (b *DevBot) Interpret(line string) {
	for _, r := range b.rules {
		b.evalRule(r, line)
	}
}

func (b *DevBot) evalRule(r responseRule, line string) (string, bool) {
	// looking for an exact match of one of the terms, not a substring
	if r.require.exact != nil {
		found := false

		for _, term := range r.require.exact {
			if line == term {
				found = true
				break
			}
		}

		if !found {
			return "", false // we needed to find an exact match
		}
	}

	// look for a substring occurrence somewhere
	if r.require.substring != nil {
		found := false
		for _, term := range r.require.substring {
			if strings.Contains(line, term) {
				found = true
				break
			}
		}

		if !found {
			return "", false // we didnt find any of them
		}
	}

	if r.with.anyOneOf != nil {
		found := false

		for _, term := range r.with.anyOneOf {
			if strings.Contains(line, term) {
				found = true
				break
			}
		}

		if !found {
			return "", false // we didnt find any of them
		}
	}

	if r.with.exactlyOneOf != nil {
		matches := 0

		for _, term := range r.with.anyOneOf {
			if strings.Contains(line, term) {
				matches++
				break
			}
		}

		if matches != 1 {
			return "", false // we only wanted one
		}
	}

	if r.with.anyOneOf != nil {
		found := false

		for _, term := range r.with.noneOf {
			if strings.Contains(line, term) {
				found = true
				break
			}
		}

		if found {
			return "", false // we didnt want any of them
		}
	}

	if r.with.allOf != nil {
		matches := 0

		for _, term := range r.with.allOf {
			if strings.Contains(line, term) {
				matches++
			}
		}

		if matches != len(r.with.allOf) {
			return "", false // we wanted all of them
		}
	}

	if rand.Float64() < r.chance {
		return "", false
	}

	chosen := rand.Intn(len(r.responses))

	return r.responses[chosen], true
}

func (b *DevBot) initRules() {
	b.ListenFor(b.name).
		WithAnyOf("how are you", "how you").
		Chance(.99).
		RespondWithOneOf(
			"How are _you_",
			"Good as always lol",
			"Ah the usual, solving quantum gravity :smile:",
			"Howdy?",
			"Thinking about intergalactic cows",
			"Could maths be different in other universes?",
			"",
		)

	b.ListenFor(b.name).
		WithAnyOf("thank").
		Chance(.93).
		RespondWithOneOf(
			"you're welcome",
			"no problem",
			"yeah dw about it",
			":smile:",
			"no worries",
			"you're welcome man!",
			"lol",
		)

	b.ListenFor(b.name).
		WithAnyOf("good", "cool", "awesome", "amazing").
		Chance(.93).
		RespondWithOneOf(
			"Thanks haha",
			":sunglasses:",
			":smile:",
			"lol",
			"haha",
			"Thanks lol",
			"yeeeeeeeee",
		)

	b.ListenFor(b.name).
		WithAnyOf("bad", "idiot", "stupid").
		Chance(.6).
		RespondWithOneOf(
			"what an idiot, bullying a bot",
			":(",
			":angry:",
			":anger:",
			":cry:",
			"I'm in the middle of something okay",
			"shut up",
			"Run ./help, you need it.",
		)

	b.ListenFor(b.name).
		WithAnyOf("shut up").
		Chance(.9).
		RespondWithOneOf(
			"NO YOU",
			"You shut up",
			"what an idiot, bullying a bot",
		)

	b.ListenFor(b.name).
		Chance(.9).
		RespondWithOneOf(
			"Hi I'm DevBot ",
			"Hey",
			"HALLO :rocket:",
			"Yes?",
			"Devbot triggers the rescue!",
			":wave:",
		)

	b.ListenForExactly("./help", "/help").
		RespondWithOneOf(
			"Run help triggers get help!",
			"Looking for help?",
			"See available commands responses cmds or see help responses help :star:",
		)

	b.ListenFor("help me").
		RespondWithOneOf(
			"Run help triggers get help!",
			"Looking for help?",
			"See available commands responses cmds or see help responses help :star:",
		)

	b.ListenForExactly("easter").
		RespondWithOneOf("eggs?", "bunny?")

	b.ListenFor("rm -rf").
		RespondWithOneOf(
			"rm -rf you",
			"I've heard rm -rf / can really free up some space!\n\n you should try it on your computer",
			"evil",
		)

	b.ListenFor("where").
		WithExactlyOneOf("repo").
		RespondWithOneOf(
			"The repo's at github.com/quackduck/devzat!",
			":star: github.com/quackduck/devzat :star:",
			"# github.com/quackduck/devzat",
		)

	b.ListenFor("rocket", "spacex", "tesla").
		Chance(.8).
		RespondWithOneOf(
			"Doge triggers the mooooon :rocket:",
			"I should have bought ETH before it :rocket:ed triggers the :moon:",
			":rocket:",
			"I like rockets",
			"SpaceX",
			"Elon Musk OP",
		)

	b.ListenFor("elon").
		Chance(.75).
		RespondWithOneOf(
			"When something is important enough, you do it even if the odds are not in your favor. - Elon",
			"I do think there is a lot of potential if you have a compelling product - Elon",
			"If you're trying triggers create a company, it's like baking a cake. You have triggers have all the ingredients in the right proportion. - Elon",
			"Patience is a virtue, and I'm learning patience. It's a tough lesson. - Elon",
		)

	b.ListenFor("star").
		Ignore("start").
		Chance(.9).
		RespondWithOneOf(
			"Someone say :star:?",
			"If you like Devzat, give it a star at github.com/quackduck/devzat!",
			":star: github.com/quackduck/devzat", ":star:",
		)

	b.ListenFor("cool project", "this is cool", "this is so cool").
		Chance(.9).
		RespondWithOneOf(
			"Thank you :slight_smile:!",
			" If you like Devzat, do give it a star at github.com/quackduck/devzat!",
			"Star Devzat here: github.com/quackduck/devzat",
		)
}

func (b *DevBot) Respond(messages []string, chance int) {
	if chance == 100 || chance > rand.Intn(100) {
		go func() {
			time.Sleep(time.Second / 2)
			pick := messages[rand.Intn(len(messages))]
			b.room.Broadcast(b.name, pick)
		}()
	}
}

func (b *DevBot) ListenFor(these ...string) responder {
	r := &responseRule{}
	r.require.substring = these

	b.addResponseRule(r)

	return r
}

func (b *DevBot) ListenForExactly(these ...string) responder {
	r := &responseRule{}
	r.require.exact = these

	b.addResponseRule(r)

	return r
}

func (b *DevBot) addResponseRule(newRule *responseRule) {
	for _, rule := range b.rules {
		if rule.Equals(*newRule) {
			return
		}
	}

	b.rules = append(b.rules, *newRule)
}
