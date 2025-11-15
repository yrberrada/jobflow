package store

import "context"

// SaveNotionPageID stores the Notion page ID for a given application.
func (s *Store) SaveNotionPageID(ctx context.Context, appID int64, notionPageID string) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE applications SET notion_page_id = ? WHERE id = ?`,
		notionPageID, appID,
	)
	return err
}
