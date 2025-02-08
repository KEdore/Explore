CREATE TABLE decisions (
    actor_user_id VARCHAR(255) NOT NULL,
    recipient_user_id VARCHAR(255) NOT NULL,
    liked_recipient BOOLEAN NOT NULL,
    PRIMARY KEY (actor_user_id, recipient_user_id)
);
