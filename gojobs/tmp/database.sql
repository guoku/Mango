create database logs DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;

CREATE TABLE `crawler_logs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(255) NOT NULL,
  `log_type` varchar(255) NOT NULL,
  `file` varchar(255) NOT NULL,
  `line` int(11) NOT NULL,
  `time` datetime NOT NULL,
  `reason` longtext NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=627 DEFAULT CHARSET=utf;
