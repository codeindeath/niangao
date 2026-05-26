package repository

import (
	"os"
	"strings"
	"testing"
)

func TestDeprecatedExperienceRepositoryMethodsAreRemoved(t *testing.T) {
	source, err := os.ReadFile("experience.go")
	if err != nil {
		t.Fatalf("read experience.go: %v", err)
	}
	experienceSource := string(source)

	for _, symbol := range []string{
		"func (r *ExperienceRepo) Create(",
		"func (r *ExperienceRepo) CreateOfficial(",
		"func (r *ExperienceRepo) ExistsByContent(",
		"func (r *ExperienceRepo) ExistsByContentExcluding(",
		"func (r *ExperienceRepo) List(",
		"func (r *ExperienceRepo) Recommend(",
		"func (r *ExperienceRepo) ListByAuthor(",
		"func (r *ExperienceRepo) ListBookmarked(",
		"func (r *ExperienceRepo) SearchByEmbedding(",
		"func (r *ExperienceRepo) UpdateReviewResult(",
	} {
		if strings.Contains(experienceSource, symbol) {
			t.Fatalf("deprecated experience repository method should be removed: %s", symbol)
		}
	}
}
