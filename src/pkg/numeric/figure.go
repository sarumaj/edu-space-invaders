package numeric

import (
	"fmt"
	"math"
)

// Circle represents a circle.
type Circle struct {
	Position Position
	Radius   Number
}

// String returns the string representation of the circle.
func (circle Circle) String() string {
	return fmt.Sprintf("Circle{%s, R:%f}", circle.Position, circle.Radius)
}

// Vertices returns the vertices of the circle by approximating it as a polygon.
func (circle Circle) Vertices() Vertices {
	var edges int
	if circle.Radius <= 3 {
		edges = 30

	} else if circle.Radius <= 10 {
		edges = circle.Radius.Int() * 10

	} else if circle.Radius <= 50 {
		edges = circle.Radius.Int() * 2

	} else {
		edges = 100

	}

	var vertices Vertices
	for i := 0; i < edges; i++ {
		angle := (2 * Pi * Number(i) / Number(edges)).Float()
		vertices = append(vertices, circle.Position.Add(Locate(math.Cos(angle), math.Sin(angle)).Mul(circle.Radius)))
	}

	return vertices
}

// Rectangle represents a rectangle.
type Rectangle [4]Position

// String returns the string representation of the rectangle.
func (rect Rectangle) String() string {
	return fmt.Sprintf("Rectangle{%v, %v, %v, %v}", rect[0], rect[1], rect[2], rect[3])
}

// Vertices returns the vertices of the rectangle.
func (rect Rectangle) Vertices() Vertices { return rect[:] }

// Triangle represents a triangle.
type Triangle [3]Position

// String returns the string representation of the triangle.
func (tri Triangle) String() string {
	return fmt.Sprintf("Triangle{%v, %v, %v}", tri[0], tri[1], tri[2])
}

// Vertices returns the vertices of the triangle.
func (tri Triangle) Vertices() Vertices { return tri[:] }

// SpaceshipPolygon represents a spaceship polygon.
type SpaceshipPolygon [7]Position

// String returns the string representation of the polygon.
func (poly SpaceshipPolygon) String() string {
	return fmt.Sprintf("SpaceshipPolygon{%v, %v, %v, %v, %v, %v, %v}", poly[0], poly[1], poly[2], poly[3], poly[4], poly[5], poly[6])
}

// Vertices returns the vertices of the spaceship polygon.
func (poly SpaceshipPolygon) Vertices() Vertices { return poly[:] }

// GetRectangularVertices calculates the vertices of the rectangle.
func GetRectangularVertices(pos Position, size Size) Rectangle {
	return Rectangle{
		Locate(pos.X, pos.Y+size.Height),            // Bottom-left
		Locate(pos.X+size.Width, pos.Y+size.Height), // Bottom-right
		Locate(pos.X+size.Width, pos.Y),             // Top-right
		pos,                                         // Top-left
	}
}

// GetSpaceshipVerticesV1 calculates the vertices of the spaceship.
// The spaceship is approximated as a triangle.
// The spaceship can face up or down.
func GetSpaceshipVerticesV1(pos Position, size Size, faceUp bool) Triangle {
	if faceUp {
		return Triangle{
			Locate(pos.X, pos.Y+size.Height),            // Bottom left
			Locate(pos.X+size.Width, pos.Y+size.Height), // Bottom right
			Locate(pos.X+size.Width/2, pos.Y),           // Top
		}
	}

	return Triangle{
		Locate(pos.X+size.Width/2, pos.Y+size.Height), // Bottom
		Locate(pos.X+size.Width, pos.Y),               // Top right
		Locate(pos.X, pos.Y),                          // Top left
	}
}

// GetSpaceshipVerticesV2 calculates the vertices of the spaceship.
// The spaceship has a main body, a head and two wings.
// The main body is a rectangle, and the wings, and the head are triangles.
// The spaceship can face up or down.
// It is more precise than GetSpaceshipVerticesV1.
func GetSpaceshipVerticesV2(pos Position, size Size, faceUp bool) SpaceshipPolygon {
	if faceUp {
		return SpaceshipPolygon{
			Locate(pos.X, pos.Y+size.Height*0.75),               // Bottom-left of left wing
			Locate(pos.X+size.Width*0.4, pos.Y+size.Height*0.8), // Bottom-left of the main body
			Locate(pos.X+size.Width*0.6, pos.Y+size.Height*0.8), // Bottom-right of the main body
			Locate(pos.X+size.Width, pos.Y+size.Height*0.75),    // Bottom-right of right wing
			Locate(pos.X+size.Width*0.6, pos.Y+size.Height*0.2), // Top-right of the main body
			Locate(pos.X+size.Width*0.5, pos.Y),                 // Top point of the tip
			Locate(pos.X+size.Width*0.4, pos.Y+size.Height*0.2), // Top-left of the main body
		}
	}

	return SpaceshipPolygon{
		Locate(pos.X+size.Width*0.4, pos.Y+size.Height*0.8), // Bottom-left of the main body
		Locate(pos.X+size.Width*0.5, pos.Y+size.Height),     // Bottom point of the tip
		Locate(pos.X+size.Width*0.6, pos.Y+size.Height*0.8), // Bottom-right of the main body
		Locate(pos.X+size.Width, pos.Y+size.Height*0.25),    // Top-right of right wing
		Locate(pos.X+size.Width*0.6, pos.Y+size.Height*0.2), // Top-right of the main body
		Locate(pos.X+size.Width*0.4, pos.Y+size.Height*0.2), // Top-left of the main body
		Locate(pos.X, pos.Y+size.Height*0.25),               // Top-left of left wing
	}
}
