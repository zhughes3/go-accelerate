package user

import "context"

const labelUserID = "userId"

func IDExtractor(ctx context.Context) map[string]any {
	userID, found := ResolveID(ctx)
	if !found {
		userID = "N/A"
	}

	return map[string]any{labelUserID: userID}
}
