-- -----------------------------------------------------
-- Table `user`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `user` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `created` DATETIME NOT NULL,
  `updated` DATETIME NULL,
  `email` VARCHAR(45) NOT NULL,
  `full_name` VARCHAR(255) NOT NULL,
  `last_name` VARCHAR(255) NOT NULL,
  `enabled` TINYINT NULL,
  `first_name` VARCHAR(255) NOT NULL,
  `profile_pic_url` MEDIUMTEXT NULL,
  `locale` VARCHAR(255) NULL,
  `nickname` VARCHAR(255) NULL,
  `verified` TINYINT NULL,
  `state` VARCHAR(255) NULL,
  `county` MEDIUMTEXT NULL,
  `sub` VARCHAR(45) NULL,
  `user_id` VARCHAR(45) NULL,
  `newsletter` TINYINT NULL,
  `beta_program` TINYINT NULL,
  `billing_status` VARCHAR(255) NULL,
  `trial_start` DATETIME NULL,
  `onboarded` TINYINT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `email_UNIQUE` (`email` ASC))
ENGINE = InnoDB;
