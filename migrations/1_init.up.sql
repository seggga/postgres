BEGIN;

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id INT GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(50) NOT NULL,
    email VARCHAR(254) UNIQUE NOT NULL,
    login VARCHAR(50) UNIQUE NOT NULL,
    birthday DATE NOT NULL,
    
    PRIMARY KEY(id),
    constraint users_valid_birthday CHECK (birthday < NOW())
);

CREATE TYPE resolution AS ENUM ('144p', '240p', '360p', '480p', '720p', '1080p');

DROP TABLE IF EXISTS videos;
CREATE TABLE videos(
    id INT GENERATED ALWAYS AS IDENTITY,
    user_id INT NOT NULL,
    location VARCHAR(255) UNIQUE NOT NULL,
    uri VARCHAR(255) UNIQUE NOT NULL,
    res resolution NOT NULL,
    caption VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (id),
    constraint videos_fk_user_id FOREIGN KEY (user_id) references users (id) on delete restrict,
    constraint videos_valid_created_at CHECK (created_at <= NOW()),
    constraint videos_valid_updated_at CHECK (updated_at <= NOW() OR updated_at is NULL)    
);


DROP TABLE IF EXISTS comments;
CREATE TABLE comments (
    id INT GENERATED ALWAYS AS IDENTITY,
    user_id INT NOT NULL,
    video_id INT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,

    PRIMARY KEY (id),
    constraint comments_fk_user_id FOREIGN KEY (user_id) references users (id) on delete restrict,
    constraint comments_fk_video_id FOREIGN KEY (video_id) references videos (id) on delete restrict,
    constraint comments_valid_created_at CHECK (created_at <= NOW())     
);


DROP TABLE IF EXISTS likes;
CREATE TABLE likes(
    user_id INT REFERENCES users (id),
    video_id INT REFERENCES videos (id),
    thumb_up BOOLEAN NOT NULL,

    PRIMARY KEY (user_id, video_id),
    constraint likes_fk_user_id FOREIGN KEY (user_id) references users (id) on delete restrict,
    constraint likes_fk_video_id FOREIGN KEY (video_id) references videos (id) on delete restrict
);

COMMIT;