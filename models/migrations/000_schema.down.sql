DROP TRIGGER IF EXISTS update_item_updated_at ON public.item;
DROP TABLE IF EXISTS public.item;
DROP TRIGGER IF EXISTS update_user_updated_at ON public."user";
DROP TABLE IF EXISTS public."user";
DROP FUNCTION IF EXISTS update_updated_at_column();