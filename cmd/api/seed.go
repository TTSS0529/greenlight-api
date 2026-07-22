package main

import (
	"errors"
	"fmt"

	"github.com/TTSS0529/greenlight-api/internal/data"
)

func (app *application) seed() error {
	fmt.Println("seeding database...")
	// Create users
	users := []struct {
		name        string
		email       string
		password    string
		permissions []string
	}{
		{
			name:     "Admin User",
			email:    "admin@example.com",
			password: "password",
			permissions: []string{
				"movies:read",
				"movies:write",
			},
		},
		{
			name:     "Reader User",
			email:    "reader@example.com",
			password: "password",
			permissions: []string{
				"movies:read",
			},
		},
		{
			name:        "Guest User",
			email:       "guest@example.com",
			password:    "password",
			permissions: []string{},
		},
	}
	for _, item := range users {
		user := &data.User{
			Name:      item.name,
			Email:     item.email,
			Activated: true,
		}
		err := user.Password.Set(item.password)
		if err != nil {
			return err
		}
		err = app.models.Users.Insert(user)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrDuplicateEmail):
				fmt.Println("user already exists:", item.email)
				continue
			default:
				return err
			}
		}
		if len(item.permissions) > 0 {
			err = app.models.Permissions.AddForUser(
				user.ID,
				item.permissions...,
			)
			if err != nil {
				return err
			}
		}
		fmt.Println("created user:", item.email)
	}
	// Create movies
	movies := []data.Movie{
		{
			Title:   "The Matrix",
			Year:    1999,
			Runtime: 136,
			Genres: []string{
				"action",
				"sci-fi",
			},
		},
		{
			Title:   "Inception",
			Year:    2010,
			Runtime: 148,
			Genres: []string{
				"action",
				"sci-fi",
			},
		},
		{
			Title:   "The Godfather",
			Year:    1972,
			Runtime: 175,
			Genres: []string{
				"crime",
				"drama",
			},
		},
	}
	for _, movie := range movies {
		err := app.models.Movies.Insert(&movie)
		if err != nil {
			return err
		}
		fmt.Println("created movie:", movie.Title)
	}
	fmt.Println("database seed completed")
	return nil
}
