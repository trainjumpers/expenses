-- +goose Up
-- +goose StatementBegin
INSERT INTO test.user (name, email, password, created_at, updated_at) VALUES ('Test user 1', 'test1@example.com', '$2a$10$.gzt.JGLbMl01.KnwoRvyuPAt0h.XtRuKTnaAmUPXF6r5P.8XP3kO', NOW(), NOW());
INSERT INTO test.user (name, email, password, created_at, updated_at) VALUES ('Test user 2', 'test2@example.com', '$2a$10$.gzt.JGLbMl01.KnwoRvyuPAt0h.XtRuKTnaAmUPXF6r5P.8XP3kO', NOW(), NOW());
INSERT INTO test.user (name, email, password, created_at, updated_at) VALUES ('Test user 3', 'test3@example.com', '$2a$10$.gzt.JGLbMl01.KnwoRvyuPAt0h.XtRuKTnaAmUPXF6r5P.8XP3kO', NOW(), NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
