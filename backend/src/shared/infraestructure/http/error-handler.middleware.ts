import { NextFunction, Request, Response } from "express";
import { DomainError } from "../../domain/DomainError";

export function errorHandlerMiddleware(
  error: Error,
  _req: Request,
  res: Response,
  _next: NextFunction
): void {
  if (error instanceof DomainError) {
    res.status(400).json({
      error: "DOMAIN_ERROR",
      message: error.message
    });

    return;
  }

  console.error(error);

  res.status(500).json({
    error: "INTERNAL_SERVER_ERROR",
    message: "Unexpected error"
  });
}