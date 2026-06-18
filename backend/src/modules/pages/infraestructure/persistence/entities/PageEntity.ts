import { Entity, PrimaryGeneratedColumn, Column, ManyToOne, OneToMany, JoinColumn, CreateDateColumn, UpdateDateColumn } from "typeorm";
import { EditionEntity } from "../../../../editions/infraestructure/persistence/entities/EditionEntity";
import { PautaEntity } from "../../../../pautas/infraestructure/persistence/entities/PautaEntity";

@Entity("pages")
export class PageEntity {
    @PrimaryGeneratedColumn()
    id!: number;

    @Column({ name: "no_pagina", type: "integer" })
    no_pagina!: number;

    @Column({ name: "edicion_id", type: "integer" })
    edicion_id!: number;

    @ManyToOne(() => EditionEntity, (edition) => edition.pages)
    @JoinColumn({ name: "edicion_id" })
    edition!: EditionEntity;

    @OneToMany(() => PautaEntity, (pauta) => pauta.page)
    pautas!: PautaEntity[];

    @CreateDateColumn({ name: "created_at" })
    created_at!: Date;

    @UpdateDateColumn({ name: "updated_at" })
    updated_at!: Date;
}
