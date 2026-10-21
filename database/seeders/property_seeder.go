package seeders

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

// Property type IDs (see property_type_seeder.go)
const (
	propertyTypeHouse    = 1
	propertyTypePetHotel = 3
)

// Pet kind accepted by a property, derived from its mock images
const (
	petKindCat   = "cat"
	petKindDog   = "dog"
	petKindMixed = "mixed"
)

type propertySeedData struct {
	PropertyTypeID int
	Title          string
	Description    string
	Address        string
	Province       string
	District       string
	Subdistrict    string
	Zipcode        string
	Latitude       string
	Longitude      string
	PetKind        string
	ImageDir       string   // folder name under assets/prop_mock
	Facilities     []string // facilities that the images clearly show
}

// Each host owns 25 properties: index 0-24 -> host1, 25-49 -> host2
var propertyHostEmails = []string{
	"host1@example.com",
	"host2@example.com",
}

const propertiesPerHost = 25

// 50 properties — one per folder in assets/prop_mock
var propertySeeds = []propertySeedData{
	// --- host1@example.com (1-25) : กรุงเทพฯ และปริมณฑล ---
	{propertyTypePetHotel, "Sunny Cat Penthouse", "โรงแรมแมวโทนสีเหลืองสดใส ห้องผนังกระจกใสทุกห้อง มีชั้นปีนป่ายและที่นอนนุ่ม ห้อง Penthouse สำหรับแมว 1-3 ตัว", "18 ซ.สุขุมวิท 49", "กรุงเทพมหานคร", "วัฒนา", "คลองตันเหนือ", "10110", "13.734012", "100.585044", petKindCat, "Property1", []string{"Air Conditioning", "Individual Pet Rooms", "Pet Suites", "24/7 Surveillance Camera"}},
	{propertyTypePetHotel, "Glass Suite Pet Hotel", "ห้องพักกระจกขนาด 2.1 ตร.ม. ทั้ง Deluxe Room, Junior Suite และ Cat Standard Room สะอาด ปลอดภัย ใจกลางกรุงเทพฯ", "99 ถ.ประชาราษฎร์บำเพ็ญ", "กรุงเทพมหานคร", "ห้วยขวาง", "ห้วยขวาง", "10310", "13.776521", "100.579012", petKindMixed, "Property2", []string{"Air Conditioning", "Individual Pet Rooms", "Pet Suites", "Comfortable Bedding"}},
	{propertyTypePetHotel, "Splash Dog Resort", "รีสอร์ตน้องหมาพร้อมสระว่ายน้ำ มีเสื้อชูชีพและเจ้าหน้าที่ดูแลขณะว่ายน้ำ ห้องพักแยกเป็นสัดส่วน", "215 ถ.รามคำแหง ซ.24", "กรุงเทพมหานคร", "บางกะปิ", "หัวหมาก", "10240", "13.755034", "100.645127", petKindDog, "Property3", []string{"Pet Swimming Pool", "Individual Pet Rooms", "Professional Pet Care Staff"}},
	{propertyTypePetHotel, "Aqua Dog Club", "อาคารสไตล์โมเดิร์นพร้อมสระว่ายน้ำในร่มสำหรับสุนัข มีคาเฟ่ให้เจ้าของนั่งรอ ที่นอนขนนุ่มทุกห้อง", "77 ถ.อ่อนนุช ซ.65", "กรุงเทพมหานคร", "ประเวศ", "ประเวศ", "10250", "13.716043", "100.693021", petKindDog, "Property4", []string{"Pet Swimming Pool", "Indoor Play Area", "Comfortable Bedding"}},
	{propertyTypePetHotel, "Capsule Cat Inn", "ห้องแคปซูลไม้ระแนงสำหรับแมว แสงไฟอบอุ่น มีบ้านไม้และชั้นปีนในทุกห้อง เหมาะกับลูกแมวและแมวโต", "35 ซ.ลาดพร้าว 15", "กรุงเทพมหานคร", "จตุจักร", "จอมพล", "10900", "13.808015", "100.564033", petKindCat, "Property5", []string{"Air Conditioning", "Individual Pet Rooms", "Comfortable Bedding"}},
	{propertyTypeHouse, "Castle Kitty House", "บ้านรับฝากแมว ห้องส่วนตัว B1-B4 พร้อมคอนโดแมว กระบะทราย และบ้านกล่องปราสาทให้น้องเล่น", "52 ซ.ลาดพร้าว 101", "กรุงเทพมหานคร", "ลาดพร้าว", "ลาดพร้าว", "10230", "13.824011", "100.605022", petKindCat, "Property6", []string{"Individual Pet Rooms", "Toys and Games", "Air Conditioning"}},
	{propertyTypePetHotel, "Pluto Cat Lounge", "โรงแรมแมวดีไซน์มินิมอล ชั้นปีนป่ายไม้วงกลม โถงกลางสำหรับเล่นรวม ห้อง Standard กว้างขวาง", "120 ถ.สาทรใต้", "กรุงเทพมหานคร", "สาทร", "ทุ่งมหาเมฆ", "10120", "13.719021", "100.540015", petKindCat, "Property7", []string{"Indoor Play Area", "Interactive Play Equipment", "Air Conditioning"}},
	{propertyTypePetHotel, "Timber Cat Studio", "ห้อง Studio ไม้โทนอบอุ่น ขนาดกว้างสำหรับแมว 1 ครอบครัว มีระบบฟอกอากาศและกล้องดูออนไลน์ 24 ชม.", "41 ถ.บางนา-ตราด ซ.14", "กรุงเทพมหานคร", "บางนา", "บางนาเหนือ", "10260", "13.668032", "100.605041", petKindCat, "Property8", []string{"24/7 Surveillance Camera", "Individual Pet Rooms", "Air Conditioning"}},
	{propertyTypeHouse, "Bearing Cat Stay", "ที่พักแมวใกล้ถนนแบริ่ง ห้องส่วนตัวพร้อมกระบะทราย ที่นอน และของเล่น เดินทางสะดวก", "21 ซ.สุขุมวิท 107", "กรุงเทพมหานคร", "บางนา", "บางนาใต้", "10260", "13.661044", "100.614023", petKindCat, "Property9", []string{"Individual Pet Rooms", "Feeding Service", "Air Conditioning"}},
	{propertyTypePetHotel, "Woodland Dog Lodge", "คอกไม้ระแนงกว้างขวาง เหมาะกับสุนัขพันธุ์กลางถึงใหญ่ พื้นไม้กันลื่น มีพื้นที่เดินเล่นในอาคาร", "88 ถ.นวมินทร์ ซ.42", "กรุงเทพมหานคร", "บึงกุ่ม", "คลองกุ่ม", "10240", "13.802031", "100.656012", petKindDog, "Property10", []string{"Individual Pet Rooms", "Indoor Play Area", "Professional Pet Care Staff"}},
	{propertyTypePetHotel, "Green Field Dog Camp", "แคมป์น้องหมาบนสนามหญ้ากว้าง ให้วิ่งเล่นอิสระทุกวัน พร้อมอาคารห้องพักปรับอากาศ", "199 ถ.ลำลูกกา คลอง 4", "ปทุมธานี", "ลำลูกกา", "คูคต", "12130", "13.955022", "100.637015", petKindDog, "prop 11", []string{"Outdoor Play Area", "Dog Park", "Air Conditioning"}},
	{propertyTypePetHotel, "Maple Wood Cat Hotel", "โรงแรมแมวตกแต่งไม้โทนอบอุ่น ล็อบบี้กว้าง ห้องพักมีช่องมองและชั้นนอนหลายระดับ", "45 ถ.รัตนาธิเบศร์", "นนทบุรี", "เมืองนนทบุรี", "บางกระสอ", "11000", "13.862013", "100.499031", petKindCat, "prop 12", []string{"Air Conditioning", "Pet Suites", "Indoor Play Area"}},
	{propertyTypePetHotel, "Arch Gate Pet Hotel", "ห้องพักประตูโค้งสีขาว แยกโซนน้องหมาและน้องแมวชัดเจน ลดความเครียดของสัตว์เลี้ยง", "12 ถ.แจ้งวัฒนะ", "นนทบุรี", "ปากเกร็ด", "ปากเกร็ด", "11120", "13.912044", "100.498012", petKindMixed, "prop 13", []string{"Individual Pet Rooms", "Air Conditioning", "First Aid Kit"}},
	{propertyTypePetHotel, "White Cube Pet Suites", "ห้องกระจกสีขาวสะอาดตา พร้อมคอนโดแมวและต้นไม้ในร่ม รับทั้งสุนัขพันธุ์เล็กและแมว", "8 ถ.กาญจนาภิเษก", "นนทบุรี", "บางบัวทอง", "โสนลอย", "11110", "13.910021", "100.415033", petKindMixed, "prop 14", []string{"Pet Suites", "Interactive Play Equipment", "Air Conditioning"}},
	{propertyTypePetHotel, "Blue Heart Dog Hotel", "โรงแรมสุนัขในโรงพยาบาลสัตว์ ห้องพักประตูกระจกสีน้ำเงิน มีสัตวแพทย์ดูแลตลอดการเข้าพัก", "300 ถ.รังสิต-นครนายก", "ปทุมธานี", "ธัญบุรี", "ประชาธิปัตย์", "12130", "13.990012", "100.628044", petKindDog, "prop 15", []string{"Professional Pet Care Staff", "First Aid Kit", "Individual Pet Rooms"}},
	{propertyTypePetHotel, "VetCare Pet Hotel", "โรงแรมสัตว์เลี้ยงภายในโรงพยาบาลสัตว์ มีห้องตู้ไม้สำหรับแมว ห้องรวมสำหรับสุนัข และสัตวแพทย์ 24 ชม.", "66 ถ.งามวงศ์วาน", "นนทบุรี", "เมืองนนทบุรี", "บางเขน", "11000", "13.845031", "100.517022", petKindMixed, "prop 16", []string{"Professional Pet Care Staff", "First Aid Kit", "24/7 Surveillance Camera"}},
	{propertyTypeHouse, "Peace Family Dog House", "บ้านรับฝากสุนัขแบบครอบครัว คอกแยกตัวพร้อมพัดลมและแอร์ ดูแลใกล้ชิดเหมือนอยู่บ้าน", "17 ถ.กาญจนาภิเษก ซ.บางใหญ่", "นนทบุรี", "บางใหญ่", "เสาธงหิน", "11140", "13.877024", "100.410011", petKindDog, "prop 17", []string{"Individual Pet Rooms", "Feeding Service", "Air Conditioning"}},
	{propertyTypeHouse, "Loft Garden Dog Stay", "พื้นที่โล่งสไตล์ลอฟต์ ผนังสวนแนวตั้ง โต๊ะให้เจ้าของนั่งเล่นกับน้องหมา บรรยากาศสบาย", "54 ถ.คลองหลวง คลอง 1", "ปทุมธานี", "คลองหลวง", "คลองหนึ่ง", "12120", "14.073015", "100.616032", petKindDog, "prop 18", []string{"Indoor Play Area", "Air Conditioning", "Bathing Service"}},
	{propertyTypeHouse, "Tiny Paws Dog House", "ห้องพักสำหรับสุนัขพันธุ์เล็ก บ้านหมาส่วนตัว ที่นอน และชามอาหารทุกห้อง", "9 ซ.ติวานนท์ 50", "นนทบุรี", "ปากเกร็ด", "บางตลาด", "11120", "13.895033", "100.527014", petKindDog, "prop 19", []string{"Individual Pet Rooms", "Comfortable Bedding", "Feeding Service"}},
	{propertyTypePetHotel, "Rainbow Pool Dog Resort", "รีสอร์ตน้องหมาพร้อมสระว่ายน้ำในร่ม ห่วงยางสายรุ้ง และพื้นที่ grooming", "128 ถ.ปทุมธานี-ลาดหลุมแก้ว", "ปทุมธานี", "เมืองปทุมธานี", "บ้านใหม่", "12000", "14.008021", "100.530043", petKindDog, "prop 20", []string{"Pet Swimming Pool", "Grooming Service", "Bathing Service"}},
	{propertyTypeHouse, "Pine Box Cat Inn", "ห้องไม้สนแบบกล่องสำหรับแมว มีกระบะทราย ชามอาหาร และเบาะนอนในตัว", "33 ถ.พระราม 6", "กรุงเทพมหานคร", "พญาไท", "สามเสนใน", "10400", "13.780011", "100.545032", petKindCat, "prop 21", []string{"Individual Pet Rooms", "Comfortable Bedding", "Feeding Service"}},
	{propertyTypeHouse, "Cozy Shelf Cat Home", "บ้านรับฝากแมวขนาดเล็ก ชั้นวางให้น้องนอนอาบแดด ห้องสะอาด เงียบสงบ", "71 ถ.พหลโยธิน ซ.52", "กรุงเทพมหานคร", "บางเขน", "อนุสาวรีย์", "10220", "13.870042", "100.596013", petKindCat, "prop 22", []string{"Indoor Play Area", "Air Conditioning", "Feeding Service"}},
	{propertyTypePetHotel, "Arch Glass Pet Hotel", "ห้องกระจกและตู้พักหลังคาโค้งสไตล์ยุโรป รับทั้งสุนัขพันธุ์เล็กและแมว เจ้าของดูแลเอง", "150 ถ.สุขุมวิท", "สมุทรปราการ", "เมืองสมุทรปราการ", "บางเมือง", "10270", "13.610024", "100.615011", petKindMixed, "prop 23", []string{"Individual Pet Rooms", "Air Conditioning", "Professional Pet Care Staff"}},
	{propertyTypePetHotel, "Ratchada Pet Hotel", "ล็อบบี้หินอ่อน ห้องกระจกสำหรับสุนัขและห้องคอนโดแมว มีทางเดินกว้างและระบบปรับอากาศทั้งอาคาร", "222 ถ.รัชดาภิเษก", "กรุงเทพมหานคร", "ดินแดง", "ดินแดง", "10400", "13.770033", "100.570024", petKindMixed, "prop 24", []string{"Pet Suites", "Air Conditioning", "24/7 Surveillance Camera"}},
	{propertyTypePetHotel, "Whisker Village", "โรงแรมแมวแสงไฟอบอุ่น ห้องไม้ระแนงพร้อมโซนปีนป่ายริมกระจก เปิดบริการ 10:00-20:00 น.", "47 ถ.หทัยราษฎร์", "กรุงเทพมหานคร", "คลองสามวา", "สามวาตะวันตก", "10510", "13.860014", "100.710035", petKindCat, "prop 25", []string{"Interactive Play Equipment", "Indoor Play Area", "Air Conditioning"}},

	// --- host2@example.com (26-50) : เชียงใหม่ ชลบุรี และภาคอีสาน ---
	{propertyTypePetHotel, "Golden Lattice Cat Hotel", "ห้องพักแมวไม้ระแนงสีทอง ชั้นสองระดับ มีมุมพักผ่อนริมหน้าต่าง และเจ้าหน้าที่เล่นกับน้องทุกวัน", "9 ถ.เชียงใหม่-หางดง", "เชียงใหม่", "หางดง", "หางดง", "50230", "18.688012", "98.919034", petKindCat, "prop 26", []string{"Individual Pet Rooms", "Toys and Games", "Professional Pet Care Staff"}},
	{propertyTypePetHotel, "White Cottage Dog Hotel", "บ้านสีขาวพร้อมสวนหน้าบ้าน บริการอาบน้ำตัดขนโดยช่างมืออาชีพ ห้องพักสุนัขปรับอากาศ", "88 ถ.เชียงใหม่-ลำพูน", "เชียงใหม่", "สารภี", "ยางเนิ้ง", "50140", "18.712023", "99.035011", petKindDog, "prop 27", []string{"Grooming Service", "Bathing Service", "Outdoor Play Area"}},
	{propertyTypePetHotel, "Paws Duo Hotel", "โรงแรมสำหรับแมวและสุนัข เพดานสูงโปร่ง ห้องกระจกพร้อมชั้นปีน มีหน้าร้านติดถนน", "130 ถ.เชียงใหม่-ดอยสะเก็ด", "เชียงใหม่", "ดอยสะเก็ด", "เชิงดอย", "50220", "18.862031", "99.130012", petKindMixed, "prop 28", []string{"Pet Suites", "Air Conditioning", "Interactive Play Equipment"}},
	{propertyTypePetHotel, "Hillside Cat Lounge", "โรงแรมแมวบรรยากาศไม้อบอุ่น ล็อบบี้มินิมอล ห้องพักมีช่องมองและชั้นนอนในตัว", "56 ถ.แม่ริม-สะเมิง", "เชียงใหม่", "แม่ริม", "ริมใต้", "50180", "18.913042", "98.944023", petKindCat, "prop 29", []string{"Pet Suites", "Air Conditioning", "Indoor Play Area"}},
	{propertyTypePetHotel, "Mint Cat Hotel", "บริการฝากเลี้ยงแมว ห้องกระจกสีน้ำเงินแยกเป็นสัดส่วน และห้องเล่นรวมพร้อมชั้นปีนป่ายขนาดใหญ่", "24 ถ.ซุปเปอร์ไฮเวย์", "เชียงใหม่", "เมืองเชียงใหม่", "หนองหอย", "50000", "18.765014", "99.012033", petKindCat, "prop 30", []string{"Indoor Play Area", "Individual Pet Rooms", "Air Conditioning"}},
	{propertyTypeHouse, "Peace Garden Family House", "บ้านเลี้ยงสุนัขแบบครอบครัว ห้องสีหวานติดหน้าต่าง มีม้านั่งในสวนให้น้องเล่นกับพี่เลี้ยง", "19 ถ.สุขุมวิท พัทยา ซ.89", "ชลบุรี", "บางละมุง", "หนองปรือ", "20150", "12.927031", "100.905012", petKindDog, "prop 31", []string{"Outdoor Play Area", "Comfortable Bedding", "Feeding Service"}},
	{propertyTypePetHotel, "Sunny Lawn Dog Park", "ห้องพักสุนัขแยกตัวพร้อมลานหญ้ากว้าง ให้น้องวิ่งเล่นรวมกันวันละ 2 รอบ", "77 ถ.นาเกลือ", "ชลบุรี", "บางละมุง", "นาเกลือ", "20150", "12.955024", "100.893041", petKindDog, "prop 32", []string{"Dog Park", "Outdoor Play Area", "Individual Pet Rooms"}},
	{propertyTypeHouse, "Stairway Cat House", "บ้านรับฝากแมว ผนังขั้นบันไดปีนป่ายถึงเพดาน ห้องพักประตูตาข่ายปลอดภัย", "40 ถ.สุขุมวิท ศรีราชา", "ชลบุรี", "ศรีราชา", "ศรีราชา", "20110", "13.174012", "100.931033", petKindCat, "prop 33", []string{"Interactive Play Equipment", "Individual Pet Rooms", "Air Conditioning"}},
	{propertyTypeHouse, "Pastel Cat Home", "บ้านแมวสีพาสเทล ห้องพักมีเตียงจิ๋ว กล่องซ่อนตัว และชั้นปีนริมหน้าต่าง", "15 ถ.เจิมจอมพล", "ชลบุรี", "ศรีราชา", "สุรศักดิ์", "20110", "13.155043", "100.948021", petKindCat, "prop 34", []string{"Toys and Games", "Comfortable Bedding", "Indoor Play Area"}},
	{propertyTypeHouse, "Countryside Dog Farm", "ฟาร์มรับฝากสุนัขบรรยากาศชนบท ลานหญ้ากว้าง ห้องนอนพร้อมเตียงไม้ยกพื้น", "102 ถ.บ้านสวน-หนองข้างคอก", "ชลบุรี", "เมืองชลบุรี", "บ้านสวน", "20000", "13.350031", "100.985014", petKindDog, "prop 35", []string{"Outdoor Play Area", "Dog Park", "Walking Trail"}},
	{propertyTypePetHotel, "Premier Kitty Hotel", "โรงแรมแมวราคาเริ่มต้นย่อมเยา ห้องพักพร้อมกระบะทราย หญ้าแมว และชั้นนอนรูปตัว V", "61 ถ.พระยาสัจจา", "ชลบุรี", "เมืองชลบุรี", "เสม็ด", "20000", "13.315022", "100.950043", petKindCat, "prop 36", []string{"Feeding Service", "Individual Pet Rooms", "Air Conditioning"}},
	{propertyTypePetHotel, "Pattaya Pets Land", "รับฝากเลี้ยงสุนัขและแมว ห้องไม้ระแนงทันสมัย มีสะพานแขวนและชั้นปีนสำหรับแมว", "250 ถ.เทพประสิทธิ์", "ชลบุรี", "บางละมุง", "หนองปรือ", "20150", "12.910013", "100.895032", petKindMixed, "prop 37", []string{"Interactive Play Equipment", "Individual Pet Rooms", "Air Conditioning"}},
	{propertyTypePetHotel, "Treehouse Cat Hotel", "โรงแรมแมวธีมบ้านต้นไม้ ห้องพักหน้าแมว และโถงปีนป่ายไม้ขนาดใหญ่กลางห้อง", "33 ถ.สุขุมวิท สัตหีบ", "ชลบุรี", "สัตหีบ", "สัตหีบ", "20180", "12.661024", "100.901011", petKindCat, "prop 38", []string{"Interactive Play Equipment", "Indoor Play Area", "Pet Suites"}},
	{propertyTypePetHotel, "Meadow Cat Playroom", "ห้องเล่นแมวธีมทุ่งหญ้า ชั้นปีนสีเหลืองและอุโมงค์ ให้น้องแมวออกกำลังกายทุกวัน", "8 ถ.บ้านบึง-แกลง", "ชลบุรี", "บ้านบึง", "บ้านบึง", "20170", "13.314033", "101.108022", petKindCat, "prop 39", []string{"Indoor Play Area", "Toys and Games", "Interactive Play Equipment"}},
	{propertyTypeHouse, "Teak Cat Retreat", "บ้านพักแมวไม้สักทั้งหลัง ห้องสองชั้น โพรงไม้และชั้นปีนรอบห้อง แสงธรรมชาติส่องถึง", "27 ถ.พานทอง-บ้านเก่า", "ชลบุรี", "พานทอง", "พานทอง", "20160", "13.458014", "101.095033", petKindCat, "prop 40", []string{"Pet Suites", "Interactive Play Equipment", "Comfortable Bedding"}},
	{propertyTypePetHotel, "Jungle Cat Lodge", "ที่พักแมวธีมป่า ต้นไม้ใบใหญ่ สะพานเชือก และกล่องไม้ติดผนังให้น้องนอนเล่น", "112 ถ.มิตรภาพ", "ขอนแก่น", "เมืองขอนแก่น", "ในเมือง", "40000", "16.432021", "102.836043", petKindCat, "prop 41", []string{"Interactive Play Equipment", "Cat Garden", "Air Conditioning"}},
	{propertyTypePetHotel, "Cat Villa Khon Kaen", "โรงแรมแมวห้องตู้ไม้สีขาว มีโปรโมชันพัก 3 คืนขึ้นไป ห้องโถงกว้างพร้อมคอนโดแมว", "45 ถ.มะลิวัลย์", "ขอนแก่น", "เมืองขอนแก่น", "ศิลา", "40000", "16.475032", "102.825014", petKindCat, "prop 42", []string{"Pet Suites", "Indoor Play Area", "Air Conditioning"}},
	{propertyTypePetHotel, "Tower Cat Hotel", "โรงแรมแมวกลางเมือง เสาปีนป่ายสูงถึงเพดาน ห้องพักกระจกกรอบน้ำเงิน", "9 ถ.ประชาสโมสร", "ขอนแก่น", "เมืองขอนแก่น", "ในเมือง", "40000", "16.440013", "102.820031", petKindCat, "prop 43", []string{"Interactive Play Equipment", "Individual Pet Rooms", "24/7 Surveillance Camera"}},
	{propertyTypePetHotel, "Blue Door Dog Hotel", "โรงแรมสุนัขในโรงพยาบาลสัตว์ ห้องพักประตูน้ำเงิน คอกสแตนเลสกว้าง สัตวแพทย์ดูแลตลอด", "200 ถ.มิตรภาพ บ้านเป็ด", "ขอนแก่น", "เมืองขอนแก่น", "บ้านเป็ด", "40000", "16.438024", "102.790012", petKindDog, "prop 44", []string{"Professional Pet Care Staff", "First Aid Kit", "Individual Pet Rooms"}},
	{propertyTypePetHotel, "Fairytale Cat Castle", "โรงแรมแมวธีมเทพนิยาย ปราสาทและต้นไม้ปีนป่าย ห้องพักตกแต่งเหมือนบ้านตุ๊กตา", "18 ถ.มะลิวัลย์ บ้านไผ่", "ขอนแก่น", "บ้านไผ่", "ในเมือง", "40110", "16.060021", "102.730044", petKindCat, "prop 45", []string{"Pet Suites", "Interactive Play Equipment", "Toys and Games"}},
	{propertyTypeHouse, "Catio Garden House", "บ้านแมวพร้อม Catio และสวนติดกระจก บ้านไม้หลังเล็กสำหรับนอน และทางเดินบนผนัง", "64 ถ.ทหาร", "อุดรธานี", "เมืองอุดรธานี", "หมากแข้ง", "41000", "17.413032", "102.787013", petKindCat, "prop 46", []string{"Cat Garden", "Interactive Play Equipment", "Individual Pet Rooms"}},
	{propertyTypePetHotel, "Glass Garden Cat Hotel", "ห้องพักแมวผนังกระจกสีเขียวอ่อน ชั้นปีนไม้ในทุกห้อง สว่าง โปร่ง ทำความสะอาดง่าย", "12 ถ.โพศรี", "อุดรธานี", "เมืองอุดรธานี", "หมากแข้ง", "41000", "17.400041", "102.800022", petKindCat, "prop 47", []string{"Individual Pet Rooms", "Interactive Play Equipment", "Air Conditioning"}},
	{propertyTypePetHotel, "Sarakham Pet Hotel", "โรงแรมสัตว์เลี้ยงแห่งมหาสารคาม ห้องพักบ้านหลังคาโค้งสีขาว รับทั้งสุนัขและแมว", "55 ถ.นครสวรรค์", "มหาสารคาม", "เมืองมหาสารคาม", "ตลาด", "44000", "16.185012", "103.300034", petKindMixed, "prop 48", []string{"Individual Pet Rooms", "Air Conditioning", "24/7 Surveillance Camera"}},
	{propertyTypePetHotel, "Steel Gate Dog Lodge", "ที่พักสุนัขคอกเหล็กกว้าง ประตูน้ำเงิน ทางเดินโล่ง ใกล้มหาวิทยาลัยมหาสารคาม", "31 ถ.ขามเรียง-ท่าขอนยาง", "มหาสารคาม", "กันทรวิชัย", "ขามเรียง", "44150", "16.247023", "103.248011", petKindDog, "prop 49", []string{"Individual Pet Rooms", "Professional Pet Care Staff", "Fire Alarm System"}},
	{propertyTypePetHotel, "Golden Light Pet Hotel", "ห้องพักกระจกไฟวอร์มไลท์ รับทั้งแมวและสุนัขพันธุ์ใหญ่ เจ้าหน้าที่ดูแลใกล้ชิดทั้งวัน", "140 ถ.เหล่านาดี", "ขอนแก่น", "เมืองขอนแก่น", "พระลับ", "40000", "16.412034", "102.865021", petKindMixed, "prop 50", []string{"Pet Suites", "Professional Pet Care Staff", "Air Conditioning"}},
}

func SeedProperties(db *gorm.DB) error {
	// Load host users
	hostUsers := make([]models.User, len(propertyHostEmails))
	for i, email := range propertyHostEmails {
		if err := db.Where("email = ?", email).First(&hostUsers[i]).Error; err != nil {
			log.Printf("Failed to find host user '%s': %v", email, err)
			return err
		}
	}

	// Load PetSizeCategories (small=code0, medium=code1, large=code2)
	var petSizeCategories []models.PetSizeCategory
	if err := db.Order("code asc").Find(&petSizeCategories).Error; err != nil {
		log.Printf("Failed to load pet size categories: %v", err)
		return err
	}

	// Load Pets (dog, cat)
	var allPets []models.Pet
	if err := db.Find(&allPets).Error; err != nil {
		log.Printf("Failed to load pets: %v", err)
		return err
	}

	// Index pets by name for easy lookup
	petsByName := make(map[string]models.Pet)
	for _, pet := range allPets {
		petsByName[pet.Name] = pet
	}

	// Room category template pool — varied price, pet type, size
	type roomTemplate struct {
		Name         string
		TotalRoom    int
		NumberOfPets int
		PriceMin     float64
		PriceMax     float64
		PetSizeIndex int      // 0=small, 1=medium, 2=large
		PetNames     []string // "dog", "cat", or both
	}
	roomPool := []roomTemplate{
		{"Standard Dog Room", 5, 2, 280, 380, 0, []string{"dog"}},      // 0
		{"Standard Cat Room", 5, 2, 280, 380, 0, []string{"cat"}},      // 1
		{"Cozy Pet Room", 4, 2, 250, 350, 0, []string{"dog", "cat"}},   // 2
		{"Deluxe Dog Suite", 3, 2, 550, 700, 1, []string{"dog"}},       // 3
		{"Deluxe Cat Suite", 3, 2, 550, 700, 1, []string{"cat"}},       // 4
		{"Family Suite", 3, 3, 750, 950, 1, []string{"dog", "cat"}},    // 5
		{"Premium Suite", 2, 1, 1000, 1400, 2, []string{"dog", "cat"}}, // 6
		{"VIP Suite", 1, 1, 1600, 2200, 2, []string{"dog", "cat"}},     // 7
		{"Large Dog Villa", 2, 2, 900, 1300, 2, []string{"dog"}},       // 8
		{"Cat Penthouse", 2, 3, 900, 1300, 2, []string{"cat"}},         // 9
	}

	// Sets of room indices from roomPool, chosen by the pet kind of the property
	roomSetsByKind := map[string][][]int{
		petKindCat:   {{1, 4}, {1, 4, 9}, {1, 9}},
		petKindDog:   {{0, 3}, {0, 3, 8}, {3, 8}},
		petKindMixed: {{2, 5, 7}, {0, 1, 6}, {1, 3, 5}, {0, 4, 6}},
	}

	// Random facility pool per pet kind (keeps cat-only places away from dog parks, etc.)
	commonFacilityNames := []string{
		"Air Conditioning", "Individual Pet Rooms", "Comfortable Bedding", "Toys and Games",
		"24/7 Surveillance Camera", "Fire Alarm System", "Emergency Exit", "First Aid Kit",
		"Professional Pet Care Staff", "Feeding Service",
	}
	facilityPoolByKind := map[string][]string{
		petKindCat:   append([]string{"Pet Suites", "Indoor Play Area", "Interactive Play Equipment", "Puzzle Feeders", "Treat Dispensers"}, commonFacilityNames...),
		petKindDog:   append([]string{"Outdoor Play Area", "Dog Park", "Walking Trail", "Ball Pit", "Bathing Service", "Grooming Service", "Treat Dispensers"}, commonFacilityNames...),
		petKindMixed: append([]string{"Pet Suites", "Indoor Play Area", "Outdoor Play Area", "Interactive Play Equipment", "Bathing Service", "Grooming Service"}, commonFacilityNames...),
	}

	today := time.Now().Truncate(24 * time.Hour)
	kindCounter := make(map[string]int)

	for i, p := range propertySeeds {
		host := hostUsers[i/propertiesPerHost]

		var property models.Property
		var isNew bool

		if err := db.Where("title = ?", p.Title).First(&property).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				lat, _ := strconv.ParseFloat(p.Latitude, 64)
				lon, _ := strconv.ParseFloat(p.Longitude, 64)
				propertyTypeID := p.PropertyTypeID

				property = models.Property{
					Title:          p.Title,
					Description:    p.Description,
					Longitude:      lon,
					Latitude:       lat,
					Address:        p.Address,
					Province:       p.Province,
					District:       p.District,
					Subdistrict:    p.Subdistrict,
					Country:        "Thailand",
					Zipcode:        p.Zipcode,
					PropertyTypeID: &propertyTypeID,
					HostID:         &host.ID,
					IsActive:       true,
					PublishedAt:    &today,
				}

				if err := db.Create(&property).Error; err != nil {
					log.Printf("Failed to create property '%s': %v", p.Title, err)
					continue
				}
				log.Printf("✓ Created property '%s' (host: %s)", p.Title, host.Email)
				isNew = true
			} else {
				log.Printf("Error finding property '%s': %v", p.Title, err)
				continue
			}
		} else {
			log.Printf("- Property '%s' already exists", p.Title)
		}

		// Property facilities: the ones shown in the images + random extras for the pet kind
		existingCount := db.Model(&property).Association("Facilities").Count()
		if existingCount == 0 {
			names := pickFacilityNames(p.Facilities, facilityPoolByKind[p.PetKind], rand.Intn(4)+3)

			var facilities []models.Facility
			if err := db.Where("name IN ?", names).Find(&facilities).Error; err == nil && len(facilities) > 0 {
				if err := db.Model(&property).Association("Facilities").Append(facilities); err != nil {
					log.Printf("Warning: Failed to assign facilities to '%s': %v", p.Title, err)
				} else if isNew {
					log.Printf("   ✓ Assigned %d facilities to '%s'", len(facilities), p.Title)
				}
			}
		}

		// Seed room categories (2 or 3 per property, matched to the pet kind)
		var existingRoomCount int64
		db.Model(&models.RoomCategory{}).Where("property_id = ?", property.ID).Count(&existingRoomCount)
		if existingRoomCount == 0 && len(petSizeCategories) > 0 {
			sets := roomSetsByKind[p.PetKind]
			selectedSet := sets[kindCounter[p.PetKind]%len(sets)]
			kindCounter[p.PetKind]++

			for _, poolIdx := range selectedSet {
				tmpl := roomPool[poolIdx]
				sizeCategory := petSizeCategories[tmpl.PetSizeIndex%len(petSizeCategories)]

				// Random price within range
				price := tmpl.PriceMin + rand.Float64()*(tmpl.PriceMax-tmpl.PriceMin)
				price = float64(int(price/10)) * 10 // round to nearest 10

				// Collect assigned pets
				var assignedPets []models.Pet
				for _, name := range tmpl.PetNames {
					if pet, ok := petsByName[name]; ok {
						assignedPets = append(assignedPets, pet)
					}
				}

				roomCat := models.RoomCategory{
					Name:              tmpl.Name,
					TotalRoom:         tmpl.TotalRoom,
					NumberOfPets:      tmpl.NumberOfPets,
					BasePrice:         price,
					PropertyID:        &property.ID,
					PetSizeCategoryID: &sizeCategory.ID,
					Pets:              assignedPets,
				}

				if err := db.Create(&roomCat).Error; err != nil {
					log.Printf("Warning: Failed to create room category '%s' for '%s': %v", tmpl.Name, p.Title, err)
					continue
				}

				// Create calendar entries for 365 days
				calendars := make([]models.RoomCategoryCalendar, 365)
				for d := 0; d < 365; d++ {
					calendars[d] = models.RoomCategoryCalendar{
						RoomCategoryID:   &roomCat.ID,
						Date:             today.AddDate(0, 0, d),
						AvailableRooms:   tmpl.TotalRoom,
						BookedRooms:      0,
						UnavailableRooms: 0,
						PriceAdjustment:  0,
					}
				}
				if err := db.Create(&calendars).Error; err != nil {
					log.Printf("Warning: Failed to create calendars for room '%s': %v", tmpl.Name, err)
				}

				// Random facilities for room category (3-5 facilities)
				roomNames := pickFacilityNames(nil, facilityPoolByKind[p.PetKind], rand.Intn(3)+3)
				var roomFacilities []models.Facility
				if err := db.Where("name IN ?", roomNames).Find(&roomFacilities).Error; err == nil && len(roomFacilities) > 0 {
					if err := db.Model(&roomCat).Association("Facilities").Append(roomFacilities); err != nil {
						log.Printf("Warning: Failed to assign facilities to room '%s': %v", tmpl.Name, err)
					}
				}
			}
			log.Printf("   ✓ Created %d room categories for '%s'", len(selectedSet), p.Title)
		}
	}

	log.Println("✓ Seeding properties completed")
	return nil
}

// pickFacilityNames returns the required names plus `extra` random names from pool (no duplicates)
func pickFacilityNames(required []string, pool []string, extra int) []string {
	seen := make(map[string]bool)
	names := make([]string, 0, len(required)+extra)
	for _, n := range required {
		if !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}

	shuffled := make([]string, len(pool))
	copy(shuffled, pool)
	rand.Shuffle(len(shuffled), func(x, y int) { shuffled[x], shuffled[y] = shuffled[y], shuffled[x] })

	for _, n := range shuffled {
		if extra == 0 {
			break
		}
		if !seen[n] {
			seen[n] = true
			names = append(names, n)
			extra--
		}
	}
	return names
}
