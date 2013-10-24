-- MySQL dump 10.13  Distrib 5.5.32, for debian-linux-gnu (x86_64)
--
-- Host: localhost    Database: guokuer
-- ------------------------------------------------------
-- Server version	5.5.32-0ubuntu0.12.04.1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `m_p_api_token`
--

DROP TABLE IF EXISTS `m_p_api_token`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `m_p_api_token` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `token` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `m_p_api_token`
--

LOCK TABLES `m_p_api_token` WRITE;
/*!40000 ALTER TABLE `m_p_api_token` DISABLE KEYS */;
INSERT INTO `m_p_api_token` VALUES (1,'d61995660774083ccb8b533024f9b8bb');
/*!40000 ALTER TABLE `m_p_api_token` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `m_p_key`
--

DROP TABLE IF EXISTS `m_p_key`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `m_p_key` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `data_key` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `m_p_key`
--

LOCK TABLES `m_p_key` WRITE;
/*!40000 ALTER TABLE `m_p_key` DISABLE KEYS */;
INSERT INTO `m_p_key` VALUES (1,'guokuer20130914');
/*!40000 ALTER TABLE `m_p_key` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `password_info`
--

DROP TABLE IF EXISTS `password_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `password_info` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `account` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `desc` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `password_info_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `password_info`
--

LOCK TABLES `password_info` WRITE;
/*!40000 ALTER TABLE `password_info` DISABLE KEYS */;
INSERT INTO `password_info` VALUES (1,'果库官方微博','dcea388a5ca4116388c210dbd34d83b36d2ad2','dcea388a5ca4116388b34585',''),(2,'微信公众平台','hi@guoku.com','guoku.com!@#','');
/*!40000 ALTER TABLE `password_info` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `password_permission`
--

DROP TABLE IF EXISTS `password_permission`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `password_permission` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `password_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `level` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `password_id` (`password_id`,`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `password_permission`
--

LOCK TABLES `password_permission` WRITE;
/*!40000 ALTER TABLE `password_permission` DISABLE KEYS */;
INSERT INTO `password_permission` VALUES (2,1,6,1),(3,1,4,1),(4,2,10,3);
/*!40000 ALTER TABLE `password_permission` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `permission`
--

DROP TABLE IF EXISTS `permission`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `permission` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `content_type_id` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `codename` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `permission`
--

LOCK TABLES `permission` WRITE;
/*!40000 ALTER TABLE `permission` DISABLE KEYS */;
INSERT INTO `permission` VALUES (1,1,'Can manage password','manage_password'),(2,2,'Can manage crawler','manage_crawler'),(3,3,'Can manage product','manage_product');
/*!40000 ALTER TABLE `permission` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `register_invitation`
--

DROP TABLE IF EXISTS `register_invitation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `register_invitation` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `token` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `expired` tinyint(1) NOT NULL,
  `issue_date` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=36 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `register_invitation`
--

LOCK TABLES `register_invitation` WRITE;
/*!40000 ALTER TABLE `register_invitation` DISABLE KEYS */;
INSERT INTO `register_invitation` VALUES (16,'21f9cdbe81370f7ee7c8ac58dd7941d3','shijun@guoku.com',1,'2013-09-16 11:47:07'),(17,'79ba665807f19aa072f031554ca2bdb8','liao@guoku.com',1,'2013-09-16 11:47:13'),(18,'4e69dc9cd52738671fac30e4349dbd9a','stxiong@guoku.com',1,'2013-09-16 11:47:21'),(19,'3595a1d78cd68066b10e7e7ea7f8f7b5','huiter@guoku.com',1,'2013-09-16 11:47:28'),(20,'859fdf2887b7f87f707fa93b073c76fb','zlu@guoku.com',1,'2013-09-16 11:47:34'),(21,'0ee1eb3fdcd099e5a27201002be00580','weizhe@guoku.com',0,'2013-09-16 11:47:40'),(22,'46d5ab53524932df1909e8148fa1e39b','songwei@guoku.com',1,'2013-09-16 11:47:44'),(23,'0bbafb7aaf5088854a692be7d5d49bc1','julia@guoku.com',1,'2013-09-16 11:47:51'),(24,'32e691ab9f5cf3886294925a4a899d94','wonderm@guoku.com',1,'2013-09-16 11:48:01'),(25,'b9448cd8398b29deb2c15db3539dacef','keffy@guoku.com',0,'2013-09-16 11:48:07'),(26,'2c76ed4d9ac4b9d2b4de31a7840d013f','jasonz@guoku.com',0,'2013-09-16 11:48:19'),(27,'2be9d8daf907b09888c97e44024fc9bd','xiaoke@guoku.com',1,'2013-09-16 11:48:41'),(29,'9fcc1b8e5a8fb7a02f33d4c592e03a4d','wenjuan@guoku.com',0,'2013-09-16 11:48:52'),(30,'9fe21ec1130f369684dcf05a90dd2b46','wangyu@guoku.com',0,'2013-09-16 11:49:59'),(31,'f9c20ee1f55f20708ba8eebb9d7bfe39','julia@guoku.com',1,'2013-09-18 17:00:32'),(32,'ed014a4744cf504f045f53fdf067a4ec','xiaoke@guoku.com',1,'2013-09-22 18:00:30'),(33,'d6d6f42e73014e68b4687343300d9c3f','zlu@guoku.com',1,'2013-09-29 18:12:33'),(34,'35b1bb4619897092e039538f797db6f5','liao@guoku.com',1,'2013-10-17 14:36:54'),(35,'e413a9091dfefda6852819d206472aae','wrq@guoku.com',1,'2013-10-22 16:20:20');
/*!40000 ALTER TABLE `register_invitation` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user`
--

DROP TABLE IF EXISTS `user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `email` varchar(64) NOT NULL,
  `password` varchar(128) NOT NULL,
  `name` varchar(64) NOT NULL,
  `nickname` varchar(64) NOT NULL,
  `last_login` datetime NOT NULL,
  `date_joined` datetime NOT NULL,
  `is_active` tinyint(1) NOT NULL,
  `is_admin` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `nickname` (`nickname`),
  KEY `user_name` (`name`),
  KEY `user_is_admin` (`is_admin`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user`
--

LOCK TABLES `user` WRITE;
/*!40000 ALTER TABLE `user` DISABLE KEYS */;
INSERT INTO `user` VALUES (1,'jasonz@guoku.com','9018f63d4346f5c1bb5778eae11e85c2e3faccba','邹振盛','Jason','2013-09-14 20:44:12','2013-09-14 20:44:12',1,1),(3,'stxiong@guoku.com','eb41ac23623a2c99b8a50c3f272330195e880a42','Stephen','XIONG','2013-09-16 12:17:02','2013-09-16 12:17:02',1,1),(4,'huiter@guoku.com','f4d2cdb2bd766efa8d98da11609d829cecd329db','huiter','huiter','2013-09-16 13:17:09','2013-09-16 13:17:09',1,0),(5,'wonderm@guoku.com','f42c35279f421219943f7da4823029831fd7c9f7','孟子豪','搓澡副教授','2013-09-16 14:30:05','2013-09-16 14:30:05',1,0),(7,'songwei@guoku.com','14d18c09788c7252fc729ec7284e6bfda597bfaa','宋尉','果库六里屯','2013-09-16 16:01:29','2013-09-16 16:01:29',1,0),(8,'shijun@guoku.com','d3e4aef5eb8ccbafacc2348a18606795fb9fefa3','周士钧','shijun','2013-09-16 19:57:03','2013-09-16 19:57:03',1,1),(9,'julia@guoku.com','dbbf0cad9d0c1ce77180445991c2049ac2daa5b8','Julia Yu','鱼柳柳','2013-09-18 17:02:25','2013-09-18 17:02:25',1,0),(10,'xiaoke@guoku.com','b3d8ed111045ff374500069202d88cae9cfc3ed4','贾小可','发炎君','2013-09-22 18:02:32','2013-09-22 18:02:32',1,0),(11,'zlu@guoku.com','9d26c558a927d28f41f815eb2f582ce9a26d6f49','zlu','5%','2013-09-29 18:13:59','2013-09-29 18:13:59',1,0),(12,'liao@guoku.com','8bc6fb476c2306fd776fd1327f3d5432fa8a1783','liao','liao','2013-10-17 14:39:03','2013-10-17 14:39:03',1,0),(13,'wrq@guoku.com','187fbefb521b491d81e2a5869b82141fb2425f6b','王瑞期','沙湖王','2013-10-22 16:24:09','2013-10-22 16:24:09',1,0);
/*!40000 ALTER TABLE `user` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_permission`
--

DROP TABLE IF EXISTS `user_permission`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_permission` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `permission_id` int(11) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_permission`
--

LOCK TABLES `user_permission` WRITE;
/*!40000 ALTER TABLE `user_permission` DISABLE KEYS */;
INSERT INTO `user_permission` VALUES (1,1,1),(2,1,2),(3,1,3),(5,3,2),(6,8,2),(7,4,2),(8,11,2),(9,12,2),(10,13,2);
/*!40000 ALTER TABLE `user_permission` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_profile`
--

DROP TABLE IF EXISTS `user_profile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_profile` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `department` varchar(255) NOT NULL,
  `title` varchar(255) DEFAULT NULL,
  `mobile` varchar(255) NOT NULL,
  `phone` varchar(255) DEFAULT NULL,
  `user_id` int(11) NOT NULL,
  `salt` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`),
  KEY `user_profile_department` (`department`),
  KEY `user_profile_mobile` (`mobile`),
  KEY `user_profile_phone` (`phone`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_profile`
--

LOCK TABLES `user_profile` WRITE;
/*!40000 ALTER TABLE `user_profile` DISABLE KEYS */;
INSERT INTO `user_profile` VALUES (1,'Engineering','','18600203270','',1,'519b65aa33c5d0140fa4b28843451e4f'),(3,'Engineering','','15811154466','',3,'83782ce84a5e870b14825f379cf489d2'),(4,'Product','','15210832621','',4,'56c391f71f1dd4f28a9941a10869cc98'),(5,'Marketing','','18911228634','',5,'bf7d719854c3778c6b8dbfb3af28392d'),(6,'Marketing','','15810457107','',6,'38e0a90974e7089f2f24be409a573a0f'),(7,'Engineering','','18311442843','',7,'8b6dda7a3445e4a2baae94a8f60b4127'),(8,'Product','','18601190069','',8,'269f27250f334869d33e64e9dd8df83e'),(9,'Marketing','','15810457107','',9,'e596451478a5fe6a8011b586750d1e28'),(10,'Other','','18612006067','',10,'c41d0312dab5a98a1f5fca9823f11612'),(11,'Engineering','','18201014250','',11,'89302df4bf8fa7db92d033bbf5b58f6d'),(12,'Operation','','13911109947','',12,'ce0deb36715151e3d8b0991cd9dfdfdc'),(13,'Engineering','','15366105880','',13,'6127a49c804bbc904ebb39c7bf8127c0');
/*!40000 ALTER TABLE `user_profile` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2013-10-23 12:33:08
