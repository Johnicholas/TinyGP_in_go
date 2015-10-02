package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "io/ioutil"
  "math/rand"
  "math"
)

type tiny_gp struct {
  fitness []float64
  pop [][]byte
  x []float64
  Minrandom, Maxrandom float64
  program []byte
  pc int
  Varnum, Fitnesscases, Randomnum int
  fbestpop, favgpop float64
  seed int
  avg_len float64  
  Targets [][]float64
}

const (
  ADD = 110
  SUB = 111
  MUL = 112
  DIV = 113
  FSET_START = ADD
  FSET_END = DIV
  MAX_LEN = 1000
  POPSIZE = 100000
  DEPTH = 5
  GENERATIONS = 100
  TSIZE = 2
  PMUT_PER_NODE = 0.05
  CROSSOVER_PROB = 0.9
)

func (self *tiny_gp)run() float64 { /* Interpreter */
  // fmt.Println("run", self.pc, self.program)
  primitive := self.program[self.pc]
  self.pc++
  if primitive < FSET_START {
    return self.x[primitive]
  }
  switch primitive {
    case ADD: return self.run() + self.run();
    case SUB: return self.run() - self.run();
    case MUL: return self.run() * self.run();
    case DIV:
      num, den := self.run(), self.run()
      if math.Abs(den) <= 0.001 {
        return num                            // Uhh....?
      } else {
        return num / den
      }
  }
  panic("should never get here")
  return 0.0
}


func (self *tiny_gp)traverse(buffer []byte, buffercount int) int {
  if buffer[buffercount] < FSET_START {
    return buffercount + 1
  }
  // fmt.Println("traverse 70")
  switch buffer[buffercount] {
    case ADD: fallthrough
    case SUB: fallthrough
    case MUL: fallthrough
    case DIV:
      return self.traverse(buffer, self.traverse(buffer, buffercount + 1));
  }
  panic("should never get here")
  return 0
}

func (self *tiny_gp)setup_fitness(fname string) {
  content, err := ioutil.ReadFile(fname)
  if err != nil {
    fmt.Println("Error: ", err)
  }
  json.Unmarshal(content, self)
}

func (self *tiny_gp)fitness_function(prog []byte) float64 {
  // fmt.Println("fitness_function: ", prog)
  fit := 0.0
  self.traverse(prog, 0)
  for i := 0; i < self.Fitnesscases; i++ {
    for j := 0; j < self.Varnum; j++ {
      self.x[j] = self.Targets[i][j]
    }
    // fmt.Println("about to run?")
    self.program = prog
    self.pc = 0
    result := self.run()
    fit += math.Abs( result - self.Targets[i][self.Varnum] )
  }
  // fmt.Println("done fitness function, returning: ", -fit)
  return -fit
}

func (self *tiny_gp)grow(pos int, max int, depth int) int {
  if pos >= max {
    // fmt.Println("grow 110")
    return -1
  }
  prim := byte(rand.Intn(2))
  if pos == 0 {
    prim = 1
  }
  if prim == 0 || depth == 0 {
    // fmt.Println("grow 118")
    prim = byte(rand.Intn(self.Varnum + self.Randomnum))
    buffer[pos] = prim
    return pos + 1
  } else {
    prim = byte(rand.Intn(FSET_END - FSET_START) + FSET_START)
    switch prim {
      case ADD: fallthrough
      case SUB: fallthrough
      case MUL: fallthrough
      case DIV:
        // fmt.Println("grow 129")
        buffer[pos] = prim
        return self.grow(self.grow(pos+1, max, depth-1), max, depth-1)
    }
  }
  panic("should never get here")
  return 0
}

func (self *tiny_gp)print_indiv(buffer []byte, buffercounter int) int {
  // fmt.Println("print_indiv?")
  if buffer[buffercounter] < FSET_START {
    if int(buffer[buffercounter]) < self.Varnum {
      fmt.Print("X", buffer[buffercounter] + 1, " ")
    } else {
      fmt.Print(self.x[buffer[buffercounter]])
    }
    return buffercounter + 1
  }
  fmt.Print("(")
  a1 := self.print_indiv( buffer, buffercounter + 1)
  switch buffer[buffercounter] {
    case ADD:
      fmt.Print(" + ")
    case SUB:
      fmt.Print(" - ")
    case MUL:
      fmt.Print(" * ")
    case DIV:
      fmt.Print(" / ")
  }
  a2 := self.print_indiv(buffer, a1)
  fmt.Print(")")
  return a2
}


var buffer [MAX_LEN]byte
func (self *tiny_gp)create_random_indiv( depth int ) []byte {
  len := self.grow(0, MAX_LEN, depth)
  for len < 0 {
    len = self.grow(0, MAX_LEN, depth)
  }
  ind := make([]byte, len)
  copy(ind, buffer[:])
  return ind
}

func (self *tiny_gp)create_random_pop(n int, depth int) [][]byte {
  pop := make([][]byte, n)
  for i := 0; i < n ; i++ {
    // fmt.Println("create_random_pop", i, n)
    pop[i] = self.create_random_indiv( depth )
    // fmt.Println("create_random_pop2", i, n)
    f :=  self.fitness_function( pop[i] )
    // fmt.Println("create_random_pop3", i, n)
    self.fitness[i] = f
  }
  return pop
}

func (self *tiny_gp)stats(fitness []float64, pop [][]byte, gen int) {
  best := rand.Intn(POPSIZE)
  node_count := 0
  self.fbestpop = fitness[best]
  self.favgpop = 0.0
  for i := 0; i < POPSIZE; i++ {
    node_count += self.traverse(pop[i], 0)
    self.favgpop += fitness[i]
    if fitness[i] > self.fbestpop {
      best = i
      self.fbestpop = fitness[i]
    }
  }
  avg_len := node_count / POPSIZE
  self.favgpop /= POPSIZE
  fmt.Println("Generation=", gen)
  fmt.Println("Avg Fitness=", -self.favgpop)
  fmt.Println("Best Fitness=", -self.fbestpop)
  fmt.Println("Avg Size=", avg_len)
  fmt.Println("Best Individual: ")
  self.print_indiv(pop[best], 0);
  fmt.Println("")
}

func (self *tiny_gp)tournament( fitness []float64, tsize int) int {
  best := rand.Intn(POPSIZE)
  fbest := -1e34
  for i := 0; i < tsize; i++ {
    competitor := rand.Intn(POPSIZE)
    if self.fitness[competitor] > fbest {
      fbest = self.fitness[competitor]
      best = competitor
    }
  }
  return best
}

func (self *tiny_gp)negative_tournament( fitness []float64, tsize int) int {
  worst := rand.Intn(POPSIZE)
  fworst := 1e34
  for i := 0; i < tsize; i++ {
    competitor := rand.Intn(POPSIZE)
    if self.fitness[competitor] < fworst {
      fworst = fitness[competitor]
      worst = competitor
    }
  }
  return worst
}

func (self *tiny_gp)crossover( parent1 []byte, parent2 []byte ) []byte {
  len1 := self.traverse( parent1, 0 )
  len2 := self.traverse( parent2, 0 )
  xo1start := rand.Intn(len1)
  xo1end := self.traverse( parent1, xo1start )
  xo2start := rand.Intn(len2)
  xo2end := self.traverse( parent2, xo2start )
  lenoff := xo1start + (xo2end - xo2start) + len1 - xo1end
  offspring := make([]byte, lenoff)
  copy(offspring[0:xo1start], parent1[0:xo1start])
  copy(offspring[xo1start:xo1start + (xo2end-xo2start)], parent2[xo2start:])
  copy(offspring[xo1start + (xo2end-xo2start):], parent1[xo1end:])
  return offspring
}

func (self *tiny_gp)mutation( parent []byte, pmut float64 ) []byte {
  len := self.traverse( parent, 0)
  parentcopy := make([]byte, len)
  copy(parentcopy, parent)
  for i := 0; i < len; i++ {
    if rand.Float64() < pmut {
      mutsite := i
      if parentcopy[mutsite] < FSET_START {
        parentcopy[mutsite] = byte(rand.Intn(self.Varnum))
      } else {
        switch parentcopy[mutsite] {
          case ADD: fallthrough
          case SUB: fallthrough
          case MUL: fallthrough
          case DIV:
            parentcopy[mutsite] = byte(rand.Intn(FSET_END - FSET_START + 1) + FSET_START)
        }
      }
    }
  }
  return parentcopy
}


func (self *tiny_gp)print_parms() {
  fmt.Println("-- TINY GP (Go version) --")
  fmt.Println("SEED=", self.seed)
  fmt.Println("MAX_LEN=", MAX_LEN)
  fmt.Println("POPSIZE=", POPSIZE)
  fmt.Println("DEPTH=", DEPTH)
  fmt.Println("CROSSOVER_PROB=", CROSSOVER_PROB)
  fmt.Println("PMUT_PER_NODE=", PMUT_PER_NODE)
  fmt.Println("MIN_RANDOM=", self.Minrandom)
  fmt.Println("MAX_RANDOM=", self.Maxrandom)
  fmt.Println("GENERATIONS=", GENERATIONS)
  fmt.Println("TSIZE=", TSIZE)
  fmt.Println("--------------------------------")
}

func new_tiny_gp(s int, fname string) *tiny_gp {
  it := new(tiny_gp)
  it.x = make([]float64, FSET_START)
  it.fbestpop, it.favgpop = 0.0, 0.0
  it.fitness = make([]float64, POPSIZE)
  it.seed = s
  if s >= 0 {
    rand.Seed(int64(s))
  }

  it.setup_fitness(fname)
  it.pop = it.create_random_pop(POPSIZE, DEPTH)
  for i := 0; i < FSET_START; i++ {
    it.x[i] = (it.Maxrandom - it.Minrandom) * rand.Float64() + it.Minrandom
  }
  return it
}

func (self *tiny_gp) evolve() {
  self.print_parms()
  self.stats( self.fitness, self.pop, 0 )
  for gen := 1; gen < GENERATIONS; gen++ {
    if self.fbestpop > -1e-5 {
      fmt.Println("PROBLEM SOLVED")
      return
    }
    for indivs := 0; indivs < POPSIZE; indivs++ {
      var newind []byte
      if rand.Float64() > CROSSOVER_PROB {
        parent1 := self.tournament( self.fitness, TSIZE )
        parent2 := self.tournament( self.fitness, TSIZE )
        newind = self.crossover( self.pop[parent1], self.pop[parent2] )
      } else {
        parent := self.tournament( self.fitness, TSIZE )
        newind = self.mutation( self.pop[parent], PMUT_PER_NODE )
      }
      newfit := self.fitness_function( newind )
      offspring := self.negative_tournament( self.fitness, TSIZE );
      self.pop[offspring] = newind;
      self.fitness[offspring] = newfit;
    }
    self.stats( self.fitness, self.pop, gen )
  }
  fmt.Println("PROBLEM *NOT* SOLVED\n");
}

func main() {
  sPtr := flag.Int("s", -1, "seed")
  fnamePtr := flag.String("fname", "problem.dat", "input filename")
  flag.Parse()
  var gp = new_tiny_gp(*sPtr, *fnamePtr)
  gp.evolve()
}
