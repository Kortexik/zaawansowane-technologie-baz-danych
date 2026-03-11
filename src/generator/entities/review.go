package entities

import (
	"benchmark/generator"
	"benchmark/types"
	"fmt"
	"time"
)

type ReviewGenerator struct {
	generator.Base
	review types.Review
}

func NewReviewGenerator(id, userID, productID int) *ReviewGenerator {
	r := generator.GenerateReview(id, userID, productID)
	return &ReviewGenerator{
		Base:   generator.NewBase("reviews", "reviews", "review", r.ID, r),
		review: r,
	}
}

func (g *ReviewGenerator) SQLInsert() string {
	r := g.review
	return fmt.Sprintf(
		"INSERT INTO reviews (id, user_id, product_id, rating, title, comment, created_at, updated_at, is_verified, helpful_votes) VALUES (%d, %d, %d, %d, '%s', '%s', '%s', '%s', %d, %d);",
		r.ID, r.UserID, r.ProductID, r.Rating, r.Title, r.Comment,
		r.CreatedAt.Format(generator.SQLTimeFormat), r.UpdatedAt.Format(generator.SQLTimeFormat),
		generator.BoolToInt(r.IsVerified), r.HelpfulVotes,
	)
}

func (g *ReviewGenerator) MongoInsert() string {
	r := g.review
	return fmt.Sprintf(
		"db.reviews.insertOne({_id: %d, userId: %d, productId: %d, rating: %d, title: '%s', comment: '%s', createdAt: ISODate('%s'), updatedAt: ISODate('%s'), isVerified: %v, helpfulVotes: %d});\n",
		r.ID, r.UserID, r.ProductID, r.Rating, r.Title, r.Comment,
		r.CreatedAt.Format(time.RFC3339), r.UpdatedAt.Format(time.RFC3339),
		r.IsVerified, r.HelpfulVotes,
	)
}
