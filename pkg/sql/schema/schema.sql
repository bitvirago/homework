CREATE TABLE IF NOT EXISTS public.tasks (
                                            id uuid NOT NULL,
                                            command text,
                                            started_at timestamp without time zone,
                                            finished_at timestamp without time zone,
                                            status text,
                                            stdout text,
                                            stderr text,
                                            exit_code integer
);