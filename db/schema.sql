CREATE TABLE IF NOT EXISTS chat_messages (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  from_user VARCHAR(255) NOT NULL,
  to_user VARCHAR(255) NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  INDEX idx_from_to (from_user, to_user),
  INDEX idx_created_at (created_at)
);
