package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Favorite struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Year    int32    `json:"year,omitempty"`
	Runtime Runtime  `json:"runtime,omitempty"`
	Genres  []string `json:"genres,omitempty"`
}

type FavoriteModel struct {
	DB *sql.DB
}

func (m FavoriteModel) Add(userID, movieID int64) error {
	query := `
INSERT INTO favorite_movies (user_id, movie_id)
VALUES ($1, $2)
ON CONFLICT (user_id, movie_id) DO NOTHING`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, movieID)

	return err
}

func (m FavoriteModel) Remove(userID, movieID int64) error {
	query := `
DELETE FROM favorite_movies
WHERE user_id = $1
AND movie_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, movieID)

	return err
}

func (m FavoriteModel) GetAll(userID int64) ([]Favorite, error) {
	query := `
SELECT
	m.id,
	m.title,
	m.year,
	m.runtime,
	m.genres
FROM movies m
INNER JOIN favorite_movies f
	ON m.id = f.movie_id
WHERE f.user_id = $1
ORDER BY f.created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	favorites := []Favorite{}
	for rows.Next() {
		var favorite Favorite

		err := rows.Scan(
			&favorite.ID,
			&favorite.Title,
			&favorite.Year,
			&favorite.Runtime,
			pq.Array(&favorite.Genres),
		)
		if err != nil {
			return nil, err
		}

		favorites = append(favorites, favorite)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return favorites, nil
}
