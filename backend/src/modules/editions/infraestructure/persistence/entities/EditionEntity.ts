import { Entity, PrimaryGeneratedColumn, Column, OneToMany, CreateDateColumn, UpdateDateColumn } from "typeorm";
import { PageEntity } from "../../../../pages/infraestructure/persistence/entities/PageEntity";

@Entity("editions")
export class EditionEntity {
    @PrimaryGeneratedColumn()
    id!: number;

    @Column({ name: "no_paginas", type: "integer" })
    no_paginas!: number;

    @Column({ name: "ancho_mm", type: "decimal", precision: 10, scale: 2 })
    ancho_mm!: number;

    @Column({ name: "alto_mm", type: "decimal", precision: 10, scale: 2 })
    alto_mm!: number;

    @Column({ name: "cuadros_ancho", type: "integer" })
    cuadros_ancho!: number;

    @Column({ name: "cuadros_alto", type: "integer" })
    cuadros_alto!: number;

    @Column({ name: "facing_pages", type: "boolean", default: false })
    facing_pages!: boolean;

    @Column({ name: "margen_superior_mm", type: "decimal", precision: 10, scale: 2, default: 0 })
    margen_superior_mm!: number;

    @Column({ name: "margen_inferior_mm", type: "decimal", precision: 10, scale: 2, default: 0 })
    margen_inferior_mm!: number;

    @Column({ name: "margen_izquierdo_mm", type: "decimal", precision: 10, scale: 2, default: 0 })
    margen_izquierdo_mm!: number;

    @Column({ name: "margen_derecho_mm", type: "decimal", precision: 10, scale: 2, default: 0 })
    margen_derecho_mm!: number;

    @OneToMany(() => PageEntity, (page) => page.edition)
    pages!: PageEntity[];

    @CreateDateColumn({ name: "created_at" })
    created_at!: Date;

    @UpdateDateColumn({ name: "updated_at" })
    updated_at!: Date;
}
