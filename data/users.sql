DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
                         `id` varchar(45) NOT NULL,
                         `username` varchar(45) DEFAULT 'tkan',
                         `email` varchar(45) DEFAULT 'tkan@tma.com.vn',
                         `url` varchar(45) DEFAULT 'tma.com.vn',
                         `phone` varchar(45) DEFAULT '0898162382',
                         `active` char(1) DEFAULT 'D',
                         `locked` int(11) DEFAULT '0',
                         `dateofbirth` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                         PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;