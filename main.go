package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/atulantonyz/pokedexcli/internal/pokecache"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func cleanInput(text string) []string {
	text = strings.ToLower(text)
	return strings.Fields(text)
}
func commandExit(config *Config, area_name string) error {
	fmt.Printf("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return nil
}
func commandHelp(config *Config, area_name string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage")
	fmt.Println("")
	for _, v := range getCommands() {
		fmt.Printf("%s: %s\n", v.name, v.description)
	}

	return nil
}

func commandMap(config *Config, area_name string) error {
	url := *config.Next
	data, ok := config.Cache.Get(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("Error creating request: %w", err)
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		config.Cache.Add(url, data)
	}
	var location_area LocationArea
	if err := json.Unmarshal(data, &location_area); err != nil {
		return err
	}
	for _, la := range location_area.Results {
		fmt.Println(la.Name)
	}
	config.Next = location_area.Next
	config.Previous = location_area.Previous
	return nil
}
func commandMapb(config *Config, area_name string) error {
	if config.Previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}
	url := *config.Previous
	data, ok := config.Cache.Get(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("Error creating request: %w", err)
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		config.Cache.Add(url, data)
	}
	var location_area LocationArea
	if err := json.Unmarshal(data, &location_area); err != nil {
		return err
	}
	for _, la := range location_area.Results {
		fmt.Println(la.Name)
	}
	config.Next = location_area.Next
	config.Previous = location_area.Previous
	return nil
}
func commandExplore(config *Config, area_name string) error {
	fmt.Println("Exploring " + area_name + "...")
	url := "https://pokeapi.co/api/v2/location-area/" + area_name
	data, ok := config.ExpCache.Get(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("Error creating request: %w", err)
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		config.ExpCache.Add(url, data)
	}
	var location_area_info LocationAreaInfo
	if err := json.Unmarshal(data, &location_area_info); err != nil {
		return err
	}
	for _, pe := range location_area_info.PokemonEncounters {
		fmt.Println(pe.Pokemon.Name)
	}

	return nil

}

func commandCatch(config *Config, pokemon string) error {
	fmt.Println("Throwing a Pokeball at " + pokemon + "...")
	url := "https://pokeapi.co/api/v2/pokemon/" + pokemon
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Error creating request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var pokemon_info PokemonInfo
	if err := json.Unmarshal(data, &pokemon_info); err != nil {
		return err
	}
	base_exp := pokemon_info.BaseExperience
	capture_probability := calc_prob(base_exp)
	//fmt.Printf("\nCapture probability: %.3f", capture_probability)
	capture_success := GenerateBernoulli(capture_probability)
	if capture_success {
		fmt.Println(pokemon + " was caught!")
		config.Pokedex[pokemon] = pokemon_info
	} else {
		fmt.Println(pokemon + " escaped!")
	}
	return nil

}

func commandInspect(config *Config, pokemon string) error {
	pokemon_info, ok := config.Pokedex[pokemon]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}
	fmt.Println("Name: " + pokemon_info.Name)
	fmt.Printf("Height: %d\n", pokemon_info.Height)
	fmt.Printf("Weight: %d\n", pokemon_info.Weight)
	fmt.Println("Stats:")
	stats := pokemon_info.Stats
	for _, stat := range stats {
		stat_name := stat.Stat.Name
		stat_val := stat.BaseStat
		fmt.Printf(" -%s: %d\n", stat_name, stat_val)
	}
	fmt.Println("Types:")
	types := pokemon_info.Types
	for _, ptype := range types {
		type_name := ptype.Type.Name
		fmt.Printf(" - %s \n", type_name)
	}
	return nil
}

func commandPokedex(config *Config, pokemon string) error {
	fmt.Println("Your Pokedex:")
	for k := range config.Pokedex {
		fmt.Println(" -" + k)
	}
	return nil

}

func calc_prob(base_exp int) float64 {
	k := 1.57569
	alpha := 0.014055
	p := k * math.Exp(-alpha*float64(base_exp))
	return p
}

func GenerateBernoulli(p float64) bool {
	if p < 0 || p > 1 {
		panic("p must be between 0 and 1")
	}

	// Seed the random number generator to ensure different results on each run
	rand.New(rand.NewSource(time.Now().UnixNano()))

	return rand.Float64() < p
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, string) error
}

func getCommands() map[string]cliCommand {
	var commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Lists next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Lists previous 20 location areas",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explores given location area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Try to catch given pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect stats of caught pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Displays pokemon filled in Pokedex",
			callback:    commandPokedex,
		},
	}
	return commands
}

func main() {

	start_url := "https://pokeapi.co/api/v2/location-area"
	new_cache := pokecache.NewCache(5 * time.Second)
	explore_cache := pokecache.NewCache(5 * time.Second)
	config := Config{
		Next:     &start_url,
		Previous: nil,
		Cache:    new_cache,
		ExpCache: explore_cache,
		Pokedex:  map[string]PokemonInfo{},
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex >")
		scanner.Scan()

		clean_words := cleanInput(scanner.Text())
		if len(clean_words) == 0 {
			continue
		}
		user_command := clean_words[0]
		area_name := strings.Join(clean_words[1:], " ")
		elem, ok := getCommands()[user_command]
		if !ok {
			fmt.Println("Unknown command")
			continue
		} else {
			err := elem.callback(&config, area_name)
			if err != nil {
				fmt.Println(err)
			}
			continue
		}
	}
}
