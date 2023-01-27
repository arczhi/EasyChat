use easychat;

create table if not exists `users`(
	`id` smallint unsigned primary key auto_increment,
    `username` varchar(12) unique not null,
    `password` varchar(70) not null
); 

alter table users auto_increment = 12126;

create table if not exists `msg`(
	`id` bigint unsigned primary key,
    `sender_id` smallint unsigned not null,
    `room_key` varchar(12) not null,
    `content` varchar(256) not null,
    `created_at` datetime not null,
    foreign key (`sender_id`) references users(`id`)
);

create table if not exists `chat_room`(
	`id` smallint unsigned primary key auto_increment,
	`room_key` varchar(12) not null,
    `member_id` smallint unsigned not null
);