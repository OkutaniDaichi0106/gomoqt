import { z } from "zod"

export const ContainerSchema = z.enum(["loc", "cmaf"])
