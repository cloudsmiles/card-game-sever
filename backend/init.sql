-- 创建数据库
CREATE DATABASE IF NOT EXISTS card_game
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE card_game;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  open_id     VARCHAR(128) DEFAULT NULL,
  union_id    VARCHAR(128) DEFAULT NULL,
  nickname    VARCHAR(64)  NOT NULL,
  avatar_url  VARCHAR(512) DEFAULT NULL,
  login_type  VARCHAR(20)  NOT NULL COMMENT 'guest | wechat_h5 | wechat_open',
  created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at  DATETIME(3)  DEFAULT NULL,
  UNIQUE KEY idx_open_id (open_id),
  KEY idx_union_id (union_id),
  KEY idx_login_type (login_type),
  KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 用户游戏积分表
CREATE TABLE IF NOT EXISTS user_scores (
  id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id      BIGINT UNSIGNED NOT NULL,
  game_type    VARCHAR(20)     NOT NULL COMMENT 'ddz | mahjong | durian',
  score        INT             NOT NULL DEFAULT 0,
  games_played INT             NOT NULL DEFAULT 0,
  games_won    INT             NOT NULL DEFAULT 0,
  updated_at   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY idx_user_game (user_id, game_type),
  CONSTRAINT fk_user_scores_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户游戏积分表';
