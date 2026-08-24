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

    @Column({ name: "ubicacion_cuadros_x", type: "integer", nullable: true })
    ubicacion_cuadros_x!: number | null;

    @Column({ name: "ubicacion_cuadros_y", type: "integer", nullable: true })
    ubicacion_cuadros_y!: number | null;

    @Column({ name: "pagina_id", type: "integer", nullable: true })
    pagina_id!: number | null;

    @Column({ name: "content_type", type: "varchar", length: 10, default: "text" })
    content_type!: string;

    @Column({ name: "image_base64", type: "text", nullable: true })
    image_base64!: string | null;

    @Column({ name: "cover_date", type: "date", nullable: true })
    cover_date!: string | null;

    @ManyToOne(() => PageEntity, (page) => page.pautas, { nullable: true })
    @JoinColumn({ name: "pagina_id" })
    page!: PageEntity | null;

    @CreateDateColumn({ name: "created_at" })
    created_at!: Date;

    @UpdateDateColumn({ name: "updated_at" })
    updated_at!: Date;
}
