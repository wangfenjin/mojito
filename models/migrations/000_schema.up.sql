CREATE TABLE public."user" (
    id uuid NOT NULL PRIMARY KEY,
    email character varying(255) NOT NULL,
    hashed_password character varying NOT NULL,
    is_active boolean NOT NULL,
    is_superuser boolean NOT NULL,
    full_name character varying(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX ix_user_email ON public."user" USING btree (email);

-- Create a trigger function to automatically update the updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add trigger to user table
CREATE TRIGGER update_user_updated_at
    BEFORE UPDATE ON public."user"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE public.item (
    id uuid NOT NULL PRIMARY KEY,
    owner_id uuid NOT NULL,
    title character varying(255) NOT NULL,
    description character varying(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_item_owner FOREIGN KEY (owner_id) REFERENCES public."user" (id) ON DELETE CASCADE
);

-- Add trigger to item table
CREATE TRIGGER update_item_updated_at
    BEFORE UPDATE ON public.item
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
