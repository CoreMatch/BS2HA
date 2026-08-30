-- phpMyAdmin SQL Dump
-- version 5.2.3
-- https://www.phpmyadmin.net/
--
-- Host: localhost
-- Generation Time: Aug 30, 2026 at 02:47 PM
-- Server version: 9.0.1
-- PHP Version: 8.4.17

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `testttt`
--

-- --------------------------------------------------------

--
-- Table structure for table `oauth2_access_tokens`
--

CREATE TABLE `oauth2_access_tokens` (
  `id` int UNSIGNED NOT NULL,
  `access_token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `client_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `scopes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `subject_type` enum('user','service') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `target_uid` int UNSIGNED DEFAULT NULL,
  `target_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `expires_at` timestamp NOT NULL,
  `revoked_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `oauth2_authorization_codes`
--

CREATE TABLE `oauth2_authorization_codes` (
  `id` int UNSIGNED NOT NULL,
  `code` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `client_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `redirect_uri` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `scopes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `code_challenge` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `code_challenge_method` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` timestamp NOT NULL,
  `consumed_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `oauth2_clients`
--

CREATE TABLE `oauth2_clients` (
  `id` int UNSIGNED NOT NULL,
  `client_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `client_secret` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` enum('public','confidential') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `grant_types` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `redirect_uris` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `scopes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_internal` tinyint(1) NOT NULL DEFAULT '0',
  `is_super` tinyint(1) NOT NULL DEFAULT '0',
  `is_active` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `oauth2_clients`
--

INSERT INTO `oauth2_clients` (`id`, `client_id`, `client_secret`, `name`, `type`, `grant_types`, `redirect_uris`, `scopes`, `is_internal`, `is_super`, `is_active`, `created_at`, `updated_at`) VALUES
(1, 'hrpauth-webui', '', 'HRPAuth WebUI', 'public', '[\"authorization_code\",\"refresh_token\"]', '[\"https://auth.mcnb.dev/oauth/callback\"]', '[\"user.read\",\"user.mojang-bind-enable\",\"profile.change-username\",\"profile.change-name\",\"texture.upload\",\"texture.delete\",\"texture.get\",\"totp.setup\",\"totp.status\"]', 1, 0, 1, '2026-08-30 14:46:25', '2026-08-30 14:46:25'),
(2, 'hrpauth-internal-super', '7480d84da11377522c88946e267a9f6010b50301ade79884c2a238aba1e5149e', 'HRPAuth Internal Super Client', 'confidential', '[\"client_credentials\"]', '[]', '[\"register.manage\",\"user.declare-email\",\"user.read.as-service\",\"user.mojang-bind-enable.as-service\",\"profile.change-username.as-service\",\"profile.change-name.as-service\",\"texture.upload.as-service\",\"texture.delete.as-service\",\"texture.get.as-service\",\"totp.setup.as-service\",\"totp.status.as-service\"]', 1, 1, 1, '2026-08-30 14:46:25', '2026-08-30 14:46:25');

-- --------------------------------------------------------

--
-- Table structure for table `oauth2_refresh_tokens`
--

CREATE TABLE `oauth2_refresh_tokens` (
  `id` int UNSIGNED NOT NULL,
  `refresh_token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `access_token_id` int UNSIGNED NOT NULL,
  `client_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `scopes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` timestamp NOT NULL,
  `revoked_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `profiles`
--

CREATE TABLE `profiles` (
  `id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `model` enum('default','slim') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'default',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `profile_properties`
--

CREATE TABLE `profile_properties` (
  `id` int NOT NULL,
  `profile_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `signature` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `schema_migrations_ha`
--

CREATE TABLE `schema_migrations_ha` (
  `version` bigint NOT NULL,
  `dirty` tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `schema_migrations_ha`
--

INSERT INTO `schema_migrations_ha` (`version`, `dirty`) VALUES
(3, 0);

-- --------------------------------------------------------

--
-- Table structure for table `sessions`
--

CREATE TABLE `sessions` (
  `id` int NOT NULL,
  `profile_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `server_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `ip` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `expires_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `tokens`
--

CREATE TABLE `tokens` (
  `id` int NOT NULL,
  `access_token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `client_token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `selected_profile_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `issued_at` bigint NOT NULL,
  `expires_in_days` int DEFAULT '15',
  `state` enum('valid','temporarily_invalid','invalid') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'valid',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `uid` int UNSIGNED NOT NULL,
  `uuid` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cbh` tinyint(1) NOT NULL DEFAULT '1' COMMENT '1=由人在 WebUI 创建, 0=由 WinnerProxy 代注册创建',
  `mojang_uuid` varchar(32) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL COMMENT 'Mojang 玩家 UUID (无连字符), 绑后唯一',
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `ip` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `permission` int NOT NULL DEFAULT '0',
  `last_sign_at` datetime DEFAULT NULL,
  `register_at` datetime DEFAULT NULL,
  `verified` tinyint(1) NOT NULL DEFAULT '0',
  `remember_token` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `username` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `mbe` tinyint(1) NOT NULL DEFAULT '0' COMMENT '1=允许同名 Mojang 玩家绑定, 0=HA 优先拒绝',
  `regip` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `totp` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `2FA` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `oauth2_access_tokens`
--
ALTER TABLE `oauth2_access_tokens`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_oauth2_access_tokens_access_token` (`access_token`),
  ADD KEY `idx_oauth2_access_tokens_client_id` (`client_id`),
  ADD KEY `idx_oauth2_access_tokens_user_id` (`user_id`),
  ADD KEY `idx_oauth2_access_tokens_target_uid` (`target_uid`),
  ADD KEY `idx_oauth2_access_tokens_target_email` (`target_email`),
  ADD KEY `idx_oauth2_access_tokens_expires_at` (`expires_at`);

--
-- Indexes for table `oauth2_authorization_codes`
--
ALTER TABLE `oauth2_authorization_codes`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_oauth2_authorization_codes_code` (`code`),
  ADD KEY `idx_oauth2_authorization_codes_client_id` (`client_id`),
  ADD KEY `idx_oauth2_authorization_codes_user_id` (`user_id`),
  ADD KEY `idx_oauth2_authorization_codes_expires_at` (`expires_at`);

--
-- Indexes for table `oauth2_clients`
--
ALTER TABLE `oauth2_clients`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_oauth2_clients_client_id` (`client_id`);

--
-- Indexes for table `oauth2_refresh_tokens`
--
ALTER TABLE `oauth2_refresh_tokens`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_oauth2_refresh_tokens_refresh_token` (`refresh_token`),
  ADD KEY `idx_oauth2_refresh_tokens_access_token_id` (`access_token_id`),
  ADD KEY `idx_oauth2_refresh_tokens_client_id` (`client_id`),
  ADD KEY `idx_oauth2_refresh_tokens_user_id` (`user_id`),
  ADD KEY `idx_oauth2_refresh_tokens_expires_at` (`expires_at`);

--
-- Indexes for table `profiles`
--
ALTER TABLE `profiles`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `profile_properties`
--
ALTER TABLE `profile_properties`
  ADD PRIMARY KEY (`id`),
  ADD KEY `profile_id` (`profile_id`);

--
-- Indexes for table `schema_migrations_ha`
--
ALTER TABLE `schema_migrations_ha`
  ADD PRIMARY KEY (`version`);

--
-- Indexes for table `sessions`
--
ALTER TABLE `sessions`
  ADD PRIMARY KEY (`id`),
  ADD KEY `profile_id` (`profile_id`),
  ADD KEY `idx_sessions_server_id` (`server_id`),
  ADD KEY `idx_sessions_expires_at` (`expires_at`);

--
-- Indexes for table `tokens`
--
ALTER TABLE `tokens`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `access_token` (`access_token`),
  ADD KEY `user_id` (`user_id`),
  ADD KEY `selected_profile_id` (`selected_profile_id`),
  ADD KEY `idx_tokens_client_token` (`client_token`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`uid`),
  ADD UNIQUE KEY `idx_uuid` (`uuid`),
  ADD UNIQUE KEY `uk_users_mojang_uuid` (`mojang_uuid`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `oauth2_access_tokens`
--
ALTER TABLE `oauth2_access_tokens`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `oauth2_authorization_codes`
--
ALTER TABLE `oauth2_authorization_codes`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `oauth2_clients`
--
ALTER TABLE `oauth2_clients`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT for table `oauth2_refresh_tokens`
--
ALTER TABLE `oauth2_refresh_tokens`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `profile_properties`
--
ALTER TABLE `profile_properties`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `sessions`
--
ALTER TABLE `sessions`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `tokens`
--
ALTER TABLE `tokens`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `oauth2_access_tokens`
--
ALTER TABLE `oauth2_access_tokens`
  ADD CONSTRAINT `oauth2_access_tokens_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`uuid`) ON DELETE CASCADE;

--
-- Constraints for table `oauth2_authorization_codes`
--
ALTER TABLE `oauth2_authorization_codes`
  ADD CONSTRAINT `oauth2_authorization_codes_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`uuid`) ON DELETE CASCADE;

--
-- Constraints for table `oauth2_refresh_tokens`
--
ALTER TABLE `oauth2_refresh_tokens`
  ADD CONSTRAINT `oauth2_refresh_tokens_ibfk_1` FOREIGN KEY (`access_token_id`) REFERENCES `oauth2_access_tokens` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `oauth2_refresh_tokens_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`uuid`) ON DELETE CASCADE;

--
-- Constraints for table `profiles`
--
ALTER TABLE `profiles`
  ADD CONSTRAINT `profiles_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`uuid`) ON DELETE CASCADE;

--
-- Constraints for table `profile_properties`
--
ALTER TABLE `profile_properties`
  ADD CONSTRAINT `profile_properties_ibfk_1` FOREIGN KEY (`profile_id`) REFERENCES `profiles` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `sessions`
--
ALTER TABLE `sessions`
  ADD CONSTRAINT `sessions_ibfk_1` FOREIGN KEY (`profile_id`) REFERENCES `profiles` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `tokens`
--
ALTER TABLE `tokens`
  ADD CONSTRAINT `tokens_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`uuid`) ON DELETE CASCADE,
  ADD CONSTRAINT `tokens_ibfk_2` FOREIGN KEY (`selected_profile_id`) REFERENCES `profiles` (`id`) ON DELETE SET NULL;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
