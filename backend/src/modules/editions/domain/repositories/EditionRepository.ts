import {Edition } from '../entities/Edition';

export interface EditionRepository {
    findById(id: number): Promise<Edition | null>;
    findAll(): Promise<Edition[]>;
    save(edition: Edition): Promise<Edition>;
}