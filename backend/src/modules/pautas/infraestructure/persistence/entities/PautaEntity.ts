import { Entity, PrimaryGeneratedColumn, Column, ManyToOne, JoinColumn, CreateDateColumn, UpdateDateColumn } from "typeorm";
import { PageEntity } from "../../../../pages/infraestructure/persistence/entities/PageEntity";

@Entity("pautas")
export class PautaEntity {
    @PrimaryGeneratedColumn()
    id!: number;

    @Column({ name: "descripcion_pauta", type: "varchar", length: 255 })
    descripcion_pauta!: string;

    @Column({ name: "cuadros_alto", type: "integer" })
    cuadros_alto!: number;

    @Column({ name: "cuadros_ancho", type: "integer" })
    cuadros_ancho!: number;

    @Column({ name: "ubicacion_cuadros_x", type: "integer" })
    ubicacion_cuadros_x!: number;

    @Column({ name: "ubicacion_cuadros_y", type: "integer" })
    ubicacion_cuadros_y!: number;

    @Column({ name: "pagina_id", type: "integer" })
    pagina_id!: number;

    @ManyToOne(() => PageEntity, (page) => page.pautas)
    @JoinColumn({ name: "pagina_id" })
    page!: PageEntity;

    @CreateDateColumn({ name: "created_at" })
    created_at!: Date;

    @UpdateDateColumn({ name: "updated_at" })
    updated_at!: Date;
}
