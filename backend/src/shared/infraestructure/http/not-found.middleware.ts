import {Request, Response} from "express";

export function notFoundMiddleware(req: Request, res: Response) {
    res.status(404).json({
        error: "Not Found",
        message: `The requested resource ${req.originalUrl} was not found on this server.`
    });
}