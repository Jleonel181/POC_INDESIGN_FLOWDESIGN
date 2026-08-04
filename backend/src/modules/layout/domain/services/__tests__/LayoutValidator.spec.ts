import { LayoutValidator } from "../LayoutValidator";
import { Edition } from "../../../../editions/domain/entities/Edition";
import { Pauta } from "../../../../pautas/domain/entities/Pauta";

describe("LayoutValidator", () => {
  const validator = new LayoutValidator();
  const edition = new Edition(1, 4, 265, 370, 5, 8, false, 10, 10, 10, 10);

  describe("validatePautaInsideGrid", () => {
    it("no lanza error si la pauta cabe", () => {
      const pauta = new Pauta(1, "Test", 2, 3, 0, 0, 1);
      expect(() => validator.validatePautaInsideGrid(edition, pauta)).not.toThrow();
    });

    it("lanza error si la pauta excede horizontalmente", () => {
      const pauta = new Pauta(1, "Test", 1, 3, 3, 0, 1); // 3+3=6 > 5
      expect(() => validator.validatePautaInsideGrid(edition, pauta)).toThrow();
    });

    it("lanza error si la pauta excede verticalmente", () => {
      const pauta = new Pauta(1, "Test", 2, 1, 0, 7, 1); // 7+2=9 > 8
      expect(() => validator.validatePautaInsideGrid(edition, pauta)).toThrow();
    });

    it("lanza error si cuadros_ancho es 0", () => {
      const pauta = new Pauta(1, "Test", 1, 0, 0, 0, 1);
      expect(() => validator.validatePautaInsideGrid(edition, pauta)).toThrow();
    });
  });

  describe("validateNoOverlap", () => {
    it("no lanza error si no hay solapamiento", () => {
      const pautas = [
        new Pauta(1, "A", 2, 3, 0, 0, 1),
        new Pauta(2, "B", 2, 2, 3, 0, 1),
      ];
      expect(() => validator.validateNoOverlap(pautas)).not.toThrow();
    });

    it("lanza error si hay solapamiento", () => {
      const pautas = [
        new Pauta(1, "A", 2, 3, 0, 0, 1),
        new Pauta(2, "B", 2, 2, 2, 0, 1), // se solapa en col 2
      ];
      expect(() => validator.validateNoOverlap(pautas)).toThrow(/superponen/);
    });

    it("no lanza error con una sola pauta", () => {
      const pautas = [new Pauta(1, "A", 4, 5, 0, 0, 1)];
      expect(() => validator.validateNoOverlap(pautas)).not.toThrow();
    });

    it("no lanza error con lista vacía", () => {
      expect(() => validator.validateNoOverlap([])).not.toThrow();
    });
  });
});
