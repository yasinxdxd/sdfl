# 📘 SDF Scene Language Documentation

This document describes the available **built-in functions** for defining Signed Distance Function (SDF) based scenes.  

Each function has a specific purpose: creating objects, transformations, operations, or scene setup.  

---

## **1. scene**

Defines the root scene.  
It contains global settings such as background, camera, and child objects.  

```c#
scene(
  background: (r, g, b),
  camera: camera(
    position: (0, 5, 10)
  ),
  children: [
    plane(
      height: 0
    )
  ]
)
```

### Parameters
- **background**: `(float, float, float)`  
  RGB color of the scene background. Values are typically between `0.0` and `1.0`.  

- **camera**: `camera`  
  A `camera` function defining the scene camera.  

- **children**: `[function, ...]`  
  A list of objects, transformations, or operations.  

---

## **2. camera**

Defines the camera position.  

```c#
camera(
  position: (x, y, z)
)
```

### Parameters
- **position**: `(float, float, float)`  
  The location of the camera in 3D space.  

**Example:**

```c#
camera(
  position: (0, 5, 10)
)
```

---

## **3. plane**

Creates an infinite plane along the XY axis.  

```c#
plane(
  height: h
)
```

### Parameters
- **height**: `float`  
  The Y-offset of the plane (distance above/below the origin).  

**Example:**

```c#
plane(
  height: 0
)
```

---

## **4. sphere**

Creates a sphere.  

```c#
sphere(
  position: (x, y, z),
  radius: r
)
```

### Parameters
- **position**: `(float, float, float)`  
  The center of the sphere.  

- **radius**: `float`  
  The radius of the sphere.  

**Example:**

```c#
sphere(
  position: (0, 2, -3),
  radius: 1.5
)
```

---

## **5. box**

Creates an axis-aligned box.  

```c#
box(
  position: (x, y, z),
  size: (sx, sy, sz)
)
```

### Parameters
- **position**: `(float, float, float)`  
  The center of the box.  

- **size**: `(float, float, float)`  
  The half-size (extents) along each axis.  

**Example:**

```c#
box(
  position: (0, 3, -5),
  size: (2, 2, 2)
)
```

---

## **6. torus**

Creates a torus (donut shape).  

```c#
torus(
  position: (x, y, z),
  radius: r,
  thickness: t
)
```

### Parameters
- **position**: `(float, float, float)`  
  The center of the torus.  

- **radius**: `float`  
  The distance from the torus center to the middle of the tube.  

- **thickness**: `float`  
  The radius of the tube.  

**Example:**

```c#
torus(
  position: (0, 5, -5),
  radius: 2,
  thickness: 1
)
```

---

## **7. rotateAround**

Rotates a child object around a given point.  

```c#
rotateAround(
  position: (x, y, z),
  rotation: (rx, ry, rz),
  child: object
)
```

### Parameters
- **position**: `(float, float, float)`  
  The pivot point for rotation.  

- **rotation**: `(float, float, float)`  
  Rotation angles in **radians** (or degrees, depending on implementation) for X, Y, Z axes.  

- **child**: `object`  
  The object or operation to rotate.  

**Example:**

```c#
rotateAround(
  position: (0, 0, -5),
  rotation: (0, 0, 0),
  child: sphere(
    position: (0, 2, -5),
    radius: 1
  )
)
```

---

## **8. smoothUnion**

Smoothly blends two child objects together.  

```c#
smoothUnion(
  child1: object,
  child2: object,
  smooth_transition: s
)
```

### Parameters
- **child1**: `object`  
  First object.  

- **child2**: `object`  
  Second object.  

- **smooth_transition**: `float`  
  Controls the softness of the blend. Higher values = smoother transition.  

**Example:**

```c#
smoothUnion(
  child1: sphere(
    position: (-1, 1, -3),
    radius: 1.2
  ),
  child2: box(
    position: (1, 1, -3),
    size: (1, 1, 1)
  ),
  smooth_transition: 0.5
)
```

---

## **9. smoothSubtraction**

Subtracts one object from another with a smooth edge.  

```c#
smoothSubtraction(
  child1: object,
  child2: object,
  smooth_transition: s
)
```

### Parameters
- **child1**: `object`  
  Base object.  

- **child2**: `object`  
  The object to subtract.  

- **smooth_transition**: `float`  
  Controls the softness of the subtraction boundary.  

**Example:**

```c#
smoothSubtraction(
  child1: box(
    position: (0, 3, -5),
    size: (2, 2, 2)
  ),
  child2: torus(
    position: (0, 5, -5),
    radius: 2,
    thickness: 1
  ),
  smooth_transition: 0.8
)
```

---

## **10. smoothIntersection**

Creates a smooth intersection between two objects.  

```c#
smoothIntersection(
  child1: object,
  child2: object,
  smooth_transition: s
)
```

### Parameters
- **child1**: `object`  
  First object.  

- **child2**: `object`  
  Second object.  

- **smooth_transition**: `float`  
  Controls the softness of the intersection region.  

**Example:**

```c#
smoothIntersection(
  child1: sphere(
    position: (0, 2, -5),
    radius: 2
  ),
  child2: box(
    position: (0, 2, -5),
    size: (2, 3, 2)
  ),
  smooth_transition: 0.6
)
```

---

# 🌍 Full Scene Examples

Here are some full examples of combining primitives, transformations, and operations into a complete scene.

---

### **Example 1 – Smooth Subtraction**

```c#
scene(
  background: (0, 0, 0),
  camera: camera(
    position: (0, 5, 10)
  ),
  children: [
    plane(
      height: 0
    ),
    rotateAround(
      position: (0, 0, -5),
      rotation: (0, 0, 0),
      child: smoothSubtraction(
        child1: box(
          position: (0, 3, -5),
          size: (2, 2, 2),
        ),
        child2: torus(
          position: (0, 5, -5),
          radius: 2,
          thickness: 1
        ),
        smooth_transition: 0.8
      )
    )
  ]
)
```

---

### **Example 2 – Smooth Union**

```c#
scene(
  background: (0, 0, 0),
  camera: camera(
    position: (0, 5, 10)
  ),
  children: [
    plane(
      height: 0
    ),
    rotateAround(
      position: (0, 0, -5),
      rotation: (0, 0, 0),
      child: smoothUnion(
        child1: box(
          position: (0, 3, -5),
          size: (2, 2, 2),
        ),
        child2: torus(
          position: (0, 5, -5),
          radius: 2,
          thickness: 1
        ),
        smooth_transition: 0.8
      )
    )
  ]
)
```

---

### **Example 3 – Smooth Intersection**

```c#
scene(
  background: (0.2, 0.2, 0.2),
  camera: camera(
    position: (0, 4, 12)
  ),
  children: [
    plane(
      height: 0
    ),
    smoothIntersection(
      child1: sphere(
        position: (-2, 2, -6),
        radius: 2
      ),
      child2: box(
        position: (0, 2, -6),
        size: (3, 2, 2)
      ),
      smooth_transition: 0.5
    )
  ]
)
```

---

### **Example 4 – Multiple Objects**

```c#
scene(
  background: (0.1, 0.1, 0.1),
  camera: camera(
    position: (0, 6, 15)
  ),
  children: [
    plane(
      height: 0
    ),
    sphere(
      position: (-3, 1, -4),
      radius: 1.5
    ),
    box(
      position: (3, 1, -6),
      size: (2, 2, 2)
    ),
    torus(
      position: (0, 2, -8),
      radius: 2,
      thickness: 0.5
    )
  ]
)
```

---

✅ With these functions, you can construct entire 3D scenes by combining **primitives, transformations, and operations**.  
